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
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"github.com/goph/emperror"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const serverHost = "localhost:5555"

type app struct {
	clientID     string
	clientSecret string
	redirectURI  string

	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider

	oauthState string

	// Does the provider use "offline_access" scope to request a refresh token
	// or does it use "access_type=offline" (e.g. Google)?
	offlineAsScope bool

	client *http.Client

	pipelineBasePath string

	banzaiCli cli.Cli

	tokenChan    chan string
	shutdownChan chan struct{}
}

func runServer(banzaiCli cli.Cli, pipelineBasePath string) (string, error) {

	issuerURL, err := url.Parse(pipelineBasePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse pipelineBasePath: %v", err)
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
		oauthState:       uuid.New().String(),
		pipelineBasePath: pipelineBasePath,
		banzaiCli:        banzaiCli,
		client: &http.Client{
			Transport: banzaiCli.HTTPTransport(),
		},
	}

	ctx := oidc.ClientContext(context.Background(), a.client)
	provider, err := oidc.NewProvider(ctx, issuerURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to query provider %q: %v", issuerURL.String(), err)
	}

	var s struct {
		// What scopes does a provider support?
		//
		// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
		ScopesSupported []string `json:"scopes_supported"`
	}
	if err := provider.Claims(&s); err != nil {
		return "", fmt.Errorf("failed to parse provider scopes_supported: %v", err)
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

	http.HandleFunc("/", a.handleLogin)
	http.HandleFunc("/login", a.handleLogin)
	http.HandleFunc("/callback", a.handleCallback)

	serverURL := fmt.Sprintf("http://%s", serverHost)
	log.Infof("Opening web browser at %s", serverURL)
	go func() {
		time.Sleep(time.Second)
		err := browser.OpenURL(serverURL)
		if err != nil {
			log.Errorf("failed to open URL: %s", err.Error())
		}
	}()

	a.shutdownChan = make(chan struct{})
	a.tokenChan = make(chan string, 1)

	server := http.Server{
		Addr: serverHost,
	}

	go a.waitShutdown(&server)

	err = server.ListenAndServe()
	<-a.shutdownChan

	select {
	case token := <-a.tokenChan:
		return token, nil
	default:
		if err == nil {
			err = errors.New("login failed")
		}
		return "", err
	}
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
	scopes = append(scopes, "openid", "profile", "email", "groups", "federated:id")
	if r.FormValue("offline_access") != "yes" {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(a.oauthState)
	} else if a.offlineAsScope {
		scopes = append(scopes, "offline_access")
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(a.oauthState)
	} else {
		authCodeURL = a.oauth2Config(scopes).AuthCodeURL(a.oauthState, oauth2.AccessTypeOffline)
	}

	http.Redirect(w, r, authCodeURL, http.StatusSeeOther)
}

func (a *app) handleCallback(w http.ResponseWriter, r *http.Request) {

	pipelineToken := ""
	defer func() {
		a.shutdownChan <- struct{}{}
		a.tokenChan <- pipelineToken
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
		if state := r.FormValue("state"); state != a.oauthState {
			http.Error(w, fmt.Sprintf("expected state %q got %q", a.oauthState, state), http.StatusBadRequest)
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

	pipelineToken, err = a.requestTokenFromPipeline(rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to request Pipeline token: %v", err), http.StatusInternalServerError)
		return
	}

	renderClosingTemplate(w)

	log.Info("successfully logged in")
}

func (a *app) requestTokenFromPipeline(rawIDToken string) (string, error) {
	pipelineURL, err := url.Parse(a.pipelineBasePath)
	if err != nil {
		return "", emperror.Wrap(err, "failed to parse Pipeline endpoint")
	}

	pipelineURL.Path = "/auth/dex/callback"

	reqBody := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(reqBody)
	writer.WriteField("id_token", rawIDToken)
	writer.Close()

	a.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
	}

	resp, err := a.client.Post(pipelineURL.String(), writer.FormDataContentType(), reqBody)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("request returned: %s", string(body))
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "user_sess" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("failed to find user_sess cookie in Pipeline response")
}

func (a *app) waitShutdown(server *http.Server) {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	// Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		log.Printf("Shutdown request (signal: %v)", sig)
	case <-a.shutdownChan:
	}

	server.Shutdown(context.Background())
	close(a.shutdownChan)
}

func renderClosingTemplate(w io.Writer) {
	w.Write([]byte(closingTemplate))
}

const closingTemplate = `
<!DOCTYPE html>
<html>
<head>
<script type="text/javascript">
	setTimeout(function () { window.location.href = "https://banzaicloud.com/docs/"; }, 5000);
</script>
</head>
<body>
	<h4>You have successfully authenticated to Pipeline. Please close this page and return to your terminal. We will redirect you to the documentation otherwise.</h4>
</body>
</html>
`
