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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"emperror.dev/errors"
	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
)

type LoginApp struct {
	baseApp

	pipelineBasePath string
}

func NewLoginApp(
	banzaiCli cli.Cli,
	idpUrl string,
	pipelineBasePath string,
) *LoginApp {
	return &LoginApp{
		baseApp: baseApp{
			OidcConfig: pipeline.OidcConfig{
				IdpUrl:       idpUrl,
				ClientId:     "banzai-cli",
				ClientSecret: "banzai-cli-secret",
			},
			redirectURI: getRedirectURL(),
			oauthState:  uuid.New().String(),
			client: &http.Client{
				Transport: banzaiCli.RoundTripper(),
			},
			banzaiCli: banzaiCli,
		},
		pipelineBasePath: pipelineBasePath,
	}
}

func (a *LoginApp) getFunctions() []handleFunction {
	return []handleFunction{
		{
			pattern: baseURLPattern,
			handler: a.handleLogin,
		},
		{
			pattern: loginURLPattern,
			handler: a.handleLogin,
		},
		{
			pattern: callbackURLPattern,
			handler: a.handleCallback,
		},
	}
}

func (a *LoginApp) handleCallback(w http.ResponseWriter, r *http.Request) {
	var pipelineToken = ""
	defer func() {
		*a.shutdownChan <- struct{}{}
		*a.responseChan <- []byte(pipelineToken)
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
	json.Indent(buff, claims, "", "  ")

	pipelineToken, err = a.requestTokenFromPipeline(rawIDToken, token.RefreshToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to request Pipeline token: %v", err), http.StatusInternalServerError)
		return
	}

	a.renderClosingTemplate(w)

	log.Info("successfully logged in")
}

func (a *LoginApp) requestTokenFromPipeline(rawIDToken string, refreshToken string) (string, error) {
	pipelineURL, err := url.Parse(a.pipelineBasePath)
	if err != nil {
		return "", errors.WrapIf(err, "failed to parse Pipeline endpoint")
	}

	pipelineURL.Path = "/auth/dex/callback"

	reqBody := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(reqBody)
	writer.WriteField("id_token", rawIDToken)
	writer.WriteField("refresh_token", refreshToken)
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

	// The old version of getting the Pipeline token, should be removed in a future release.
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "user_sess" {
			return cookie.Value, nil
		}
	}

	sessionToken := resp.Header.Get("Authorization")
	if sessionToken != "" {
		return sessionToken, nil
	}

	return "", fmt.Errorf("failed to find Authorization header in Pipeline response")
}

