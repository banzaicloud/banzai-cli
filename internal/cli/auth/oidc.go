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
	"net/http"

	"emperror.dev/errors"
	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	k8sClient "k8s.io/client-go/tools/clientcmd"
	k8sClientApi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
)

type OidcConfigDownloadApp struct {
	baseApp

	clusterID     int32
	clusterConfig pipeline.ClusterConfig
}

type claim struct {
	Email string `json:"email"`
}

func NewOIDCConfigApp(
	banzaiCli cli.Cli,
	clusterID int32,
	clusterConfig pipeline.ClusterConfig,
	config pipeline.OidcConfig,
) *OidcConfigDownloadApp {
	return &OidcConfigDownloadApp{
		baseApp: baseApp{
			OidcConfig:  config,
			redirectURI: getRedirectURL(),
			oauthState:  uuid.New().String(),
			client: &http.Client{
				Transport: banzaiCli.RoundTripper(),
			},
			banzaiCli: banzaiCli,
		},
		clusterID:     clusterID,
		clusterConfig: clusterConfig,
	}
}

type handleFunction struct {
	pattern string
	handler func(http.ResponseWriter, *http.Request)
}

func (a *OidcConfigDownloadApp) getFunctions() []handleFunction {
	return []handleFunction{
		{
			pattern: baseURLPattern,
			handler: a.handleLogin,
		},
		{
			pattern: callbackURLPattern,
			handler: a.handleDexCallback,
		},
	}
}

func (a *OidcConfigDownloadApp) handleDexCallback(w http.ResponseWriter, r *http.Request) {
	var (
		token      *oauth2.Token
		err        error
		oidcConfig []byte
	)

	defer func() {
		*a.shutdownChan <- struct{}{}
		*a.responseChan <- oidcConfig
	}()

	ctx := oidc.ClientContext(r.Context(), a.client)

	switch r.Method {
	case "GET":
		// Authorization redirect callback from OAuth2 auth flow.
		if errMsg := r.FormValue("error"); errMsg != "" {
			http.Error(w, fmt.Sprintf("%s: %s", errMsg, r.FormValue("error_description")), http.StatusBadRequest)
			return
		}
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, fmt.Sprintf("no code in request: %s", r.Form), http.StatusBadRequest)
			return
		}
		stateRaw := r.FormValue("state")
		if stateRaw == "" {
			http.Error(w, fmt.Sprintf("no state in request: %s", r.Form), http.StatusBadRequest)
			return
		}

		if stateRaw != a.oauthState {
			http.Error(w, "invalid state", http.StatusBadRequest)
		}

		oauth2Config := a.oauth2Config(nil)

		token, err = oauth2Config.Exchange(ctx, code)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get token: %s", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("method not implemented: %s", r.Method), http.StatusBadRequest)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "no id_token in token response", http.StatusInternalServerError)
		return
	}

	verifier := a.provider.Verifier(&oidc.Config{ClientID: a.ClientId})

	idToken, err := verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify ID token: %s", err), http.StatusInternalServerError)
		return
	}

	var claims claim
	err = idToken.Claims(&claims)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse claims: %s", err), http.StatusInternalServerError)
		return
	}

	oidcConfig, err = a.generateKubeConfig(r.Context(), claims, rawIDToken, token.RefreshToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate kubeconfig: %s", err), http.StatusInternalServerError)
		return
	}

	a.renderClosingTemplate(w)
}

func (a *OidcConfigDownloadApp) generateKubeConfig(ctx context.Context, claims claim, IDToken, refreshToken string) ([]byte, error) {
	config, err := k8sClient.Load([]byte(a.clusterConfig.Data))
	if err != nil {
		return nil, errors.WrapIf(err, "failed to convert pipeline k8s config to k8s client config")
	}

	authInfo := k8sClientApi.NewAuthInfo()

	authInfo.AuthProvider = &k8sClientApi.AuthProviderConfig{
		Name: "oidc",
		Config: map[string]string{
			"client-id":      a.ClientId,
			"client-secret":  a.ClientSecret,
			"id-token":       IDToken,
			"refresh-token":  refreshToken,
			"idp-issuer-url": a.IdpUrl,
		},
	}

	config.AuthInfos = map[string]*k8sClientApi.AuthInfo{claims.Email: authInfo}

	currentContext := config.Contexts[config.CurrentContext]
	currentContext.AuthInfo = claims.Email

	newCurrentContext := fmt.Sprint(claims.Email, "@", currentContext.Cluster)
	config.Contexts[newCurrentContext] = currentContext

	delete(config.Contexts, config.CurrentContext)

	config.CurrentContext = newCurrentContext

	return k8sClient.Write(*config)
}
