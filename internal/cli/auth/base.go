// Copyright © 2020 Banzai Cloud
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

package auth

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"emperror.dev/errors"
	"github.com/coreos/go-oidc"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
)

const (
	serverHost = "localhost:5555"

	baseURLPattern     = "/"
	loginURLPattern    = "/login"
	callbackURLPattern = "/callback"
)

type Server interface {
	init() error
	getFunctions() []handleFunction

	chanSetter
}

type chanSetter interface {
	setChans(shutdownChan *chan struct{}, responseChan *chan []byte)
}

type app struct {
	responseChan chan []byte
	shutdownChan chan struct{}
}

type baseApp struct {
	pipeline.OidcConfig
	redirectURI string

	oauthState string
	// Does the provider use "offline_access" scope to request a refresh token
	// or does it use "access_type=offline" (e.g. Google)?
	offlineAsScope bool

	client *http.Client

	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider

	banzaiCli cli.Cli

	shutdownChan *chan struct{}
	responseChan *chan []byte
}

func (a *baseApp) init() error {
	ctx := oidc.ClientContext(context.Background(), a.client)
	provider, err := oidc.NewProvider(ctx, a.IdpUrl)
	if err != nil {
		return fmt.Errorf("failed to query provider %q: %v", a.IdpUrl, err)
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
	a.verifier = provider.Verifier(&oidc.Config{ClientID: a.ClientId})

	return nil
}

func (a *baseApp) setChans(shutdownChan *chan struct{}, responseChan *chan []byte) {
	a.shutdownChan = shutdownChan
	a.responseChan = responseChan
}

func (a *baseApp) handleLogin(w http.ResponseWriter, r *http.Request) {
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
	scopes = append(scopes, oidc.ScopeOpenID, "profile", "email", "groups", "federated:id", oidc.ScopeOfflineAccess)
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

func (a *baseApp) oauth2Config(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.ClientId,
		ClientSecret: a.ClientSecret,
		Endpoint:     a.provider.Endpoint(),
		Scopes:       scopes,
		RedirectURL:  a.redirectURI,
	}
}

func (a *baseApp) renderClosingTemplate(w io.Writer) {
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

func RunAuthServer(authServer Server) ([]byte, error) {
	a := app{
		shutdownChan: make(chan struct{}),
		responseChan: make(chan []byte, 1),
	}

	authServer.setChans(&a.shutdownChan, &a.responseChan)

	if err := authServer.init(); err != nil {
		return nil, errors.WrapIf(err, "failed to init oidc app")
	}

	for _, hf := range authServer.getFunctions() {
		http.HandleFunc(hf.pattern, hf.handler)
	}

	serverURL := fmt.Sprintf("http://%s", serverHost)
	log.Infof("Opening web browser at %s", serverURL)
	go func() {
		time.Sleep(time.Second)
		err := browser.OpenURL(serverURL)
		if err != nil {
			log.Errorf("failed to open URL: %s", err.Error())
		}
	}()

	server := http.Server{
		Addr: serverHost,
	}

	go a.waitShutdown(&server)

	err := server.ListenAndServe()
	<-a.shutdownChan

	select {
	case response := <-a.responseChan:
		return response, nil
	default:
		if err == nil {
			err = errors.New("login failed")
		}
		return nil, err
	}
}

func getRedirectURL() string {
	return fmt.Sprintf("http://%s%s", serverHost, callbackURLPattern)
}
