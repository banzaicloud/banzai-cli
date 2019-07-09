// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package login

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/coreos/go-oidc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const exampleAppState = "I wish to wash my irish wristwatch"

type app struct {
	clientID     string
	clientSecret string
	redirectURI  string

	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider

	// Does the provider use "offline_access" scope to request a refresh token
	// or does it use "access_type=offline" (e.g. Google)?
	offlineAsScope bool

	client *http.Client

	pipelineBasePath string

	banzaiCli cli.Cli

	shutdownChan chan bool
}

func runServer(banzaiCli cli.Cli, pipelineBasePath string) error {

	issuerURL, err := url.Parse(pipelineBasePath)
	if err != nil {
		return fmt.Errorf("failed to parse pipelineBasePath: %v", err)
	}

	// detect localhost setup, and derive the issuer URL
	if issuerURL.Port() == "9090" {
		issuerURL.Host = issuerURL.Hostname() + ":5556"
	}
	issuerURL.Path = "/dex"

	a := app{
		redirectURI:      "http://localhost:5555/callback",
		clientID:         "banzai-cli",
		clientSecret:     "banzai-cli-secret",
		pipelineBasePath: pipelineBasePath,
		banzaiCli:        banzaiCli,
	}

	if a.client == nil {
		a.client = http.DefaultClient
	}

	// TODO(ericchiang): Retry with backoff
	ctx := oidc.ClientContext(context.Background(), a.client)
	provider, err := oidc.NewProvider(ctx, issuerURL.String())
	if err != nil {
		return fmt.Errorf("failed to query provider %q: %v", issuerURL.String(), err)
	}

	var s struct {
		// What scopes does a provider support?
		//
		// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := provider.Claims(&s); err != nil {
		return fmt.Errorf("failed to parse provider scopes_supported: %v", err)
	}

	if len(s.ScopesSupported) == 0 {
		// scopes_supported is a "RECOMMENDED" discovery claim, not a required
		// one. If missing, assume that the provider follows the spec and has
		// an "offline_access" scope.
		a.offlineAsScope = true
	} else {
		// See if scopes_supported has the "offline_access" scope.
		a.offlineAsScope = func() bool {
			for _, scope := range s.ScopesSupported {
				if scope == oidc.ScopeOfflineAccess {
					return true
				}
			}
			return false
		}()
	}

	a.provider = provider
	a.verifier = provider.Verifier(&oidc.Config{ClientID: a.clientID})

	a.shutdownChan = make(chan bool)

	http.HandleFunc("/", a.handleLogin)
	http.HandleFunc("/login", a.handleLogin)
	http.HandleFunc("/callback", a.handleCallback)

	go open("http://127.0.0.1:5555")
	go a.waitShutdown()
	return http.ListenAndServe("127.0.0.1:5555", nil)
}

func (a *app) oauth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		Endpoint:     a.provider.Endpoint(),
		Scopes:       scopes,
		RedirectURL:  a.redirectURI,
	}
}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	scopes := make([]string, 4)
	if extraScopes := r.FormValue("extra_scopes"); extraScopes != "" {
		scopes = strings.Split(extraScopes, " ")
	}
	var clients []string
	if crossClients := r.FormValue("cross_client"); crossClients != "" {
		clients = strings.Split(crossClients, " ")
	}
	for _, client := range clients {
		scopes = append(scopes, "audience:server:client_id:"+client)
	}

	authCodeURL := ""
	scopes = append(scopes, "openid", "profile", "email", "groups")
	if r.FormValue("offline_access") != "yes" {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(exampleAppState)
	} else if a.offlineAsScope {
		scopes = append(scopes, "offline_access")
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(exampleAppState)
	} else {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(exampleAppState, oauth2.AccessTypeOffline)
	}

	http.Redirect(w, r, authCodeURL, http.StatusSeeOther)
}

func (a *app) handleCallback(w http.ResponseWriter, r *http.Request) {

	defer func() {
		a.shutdownChan <- true
	}()

	var (
		err   error
		token *oauth2.Token
	)

	ctx := oidc.ClientContext(r.Context(), a.client)
	oauth2Config := a.oauth2Config(nil)
	switch r.Method {
	case http.MethodGet:
		// Authorization redirect callback from OAuth2 auth flow.
		if errMsg := r.FormValue("error"); errMsg != "" {
			http.Error(w, errMsg+": "+r.FormValue("error_description"), http.StatusBadRequest)
			return
		}
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, fmt.Sprintf("no code in request: %q", r.Form), http.StatusBadRequest)
			return
		}
		if state := r.FormValue("state"); state != exampleAppState {
			http.Error(w, fmt.Sprintf("expected state %q got %q", exampleAppState, state), http.StatusBadRequest)
			return
		}
		token, err = oauth2Config.Exchange(ctx, code)
	case http.MethodPost:
		// Form request from frontend to refresh a token.
		refresh := r.FormValue("refresh_token")
		if refresh == "" {
			http.Error(w, fmt.Sprintf("no refresh_token in request: %q", r.Form), http.StatusBadRequest)
			return
		}
		t := &oauth2.Token{
			RefreshToken: refresh,
			Expiry:       time.Now().Add(-time.Hour),
		}
		token, err = oauth2Config.TokenSource(ctx, t).Token()
	default:
		http.Error(w, fmt.Sprintf("method not implemented: %s", r.Method), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get token: %v", err), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in token response", http.StatusInternalServerError)
		return
	}

	idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify ID token: %v", err), http.StatusInternalServerError)
		return
	}

	var claims json.RawMessage
	idToken.Claims(&claims)

	buff := new(bytes.Buffer)
	json.Indent(buff, []byte(claims), "", "  ")

	pipelineURL, err := url.Parse(a.pipelineBasePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse Pipeline URL: %v", err), http.StatusInternalServerError)
		return
	}

	pipelineURL.Path = "/auth/dex/callback"

	reqBody := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(reqBody)
	writer.WriteField("id_token", rawIDToken)
	writer.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Post(pipelineURL.String(), writer.FormDataContentType(), reqBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Pipeline token: %v", err), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Failed to create Pipeline token:\n%s", string(body)), http.StatusInternalServerError)
		return
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "user_sess" {
			a.banzaiCli.Context().SetToken(cookie.Value)
			break
		}
	}

	renderClosingTemplate(w)

	log.Info("successfully logged in")
}

func open(url string) error {

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func (a *app) waitShutdown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	// Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		log.Printf("Shutdown request (signal: %v)", sig)
		os.Exit(0)
	case <-a.shutdownChan:
		os.Exit(0)
	}
}

func renderClosingTemplate(w io.Writer) {
	w.Write([]byte(closingTemplate))
}

const closingTemplate = `
<head>
<style>
/* make pre wrap */
pre {
white-space: pre-wrap;       /* css-3 */
white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
white-space: -pre-wrap;      /* Opera 4-6 */
white-space: -o-pre-wrap;    /* Opera 7 */
word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
</style>
</head>
<script type="text/javascript">
	setTimeout(function () { window.location.href = "https://beta.banzaicloud.io/docs/"; }, 5000);
</script>
<body>
	<h4>You have successfully authenticated to Pipeline. Please close this page and return to your terminal. We will redicret you to the documentation otherwise.</h4>
</body>
</html>
`
