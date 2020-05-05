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
	"fmt"
	"net/http"
	"net/url"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/auth"
)

func runServer(banzaiCli cli.Cli, pipelineBasePath string) (string, error) {
	baseURL, err := url.Parse(pipelineBasePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse pipelineBasePath: %v", err)
	}

	issuerURL, err := getIdPURL(baseURL)
	if err != nil {
		return "", errors.WrapIf(err, "failed to get IdP url")
	}

	lApp := auth.NewLoginApp(banzaiCli, issuerURL, pipelineBasePath)
	tokenBytes, err := auth.RunAuthServer(lApp)
	if err != nil {
		return "", errors.WrapIf(err, "login failed")
	}

	return string(tokenBytes), nil
}

func getIdPURL(baseURL *url.URL) (string, error) {
	redirectUrl := fmt.Sprintf("%s://%s", baseURL.Scheme, baseURL.Hostname())
	port := baseURL.Port()
	if port != "" {
		redirectUrl = fmt.Sprintf("%s:%s", redirectUrl, port)
	}

	// get issuerURL from header of redirect
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	resp, err := client.Get(fmt.Sprintf("%s/auth/dex/login", redirectUrl))
	if err != nil {
		return "", errors.WrapIf(err, "failed to redirect login")
	}
	var location = resp.Header.Get("Location")
	issuerURL, err := url.Parse(location)
	if err != nil {
		return "", errors.WrapIf(err, "failed to get issuer url")
	}

	var finalUrl = fmt.Sprintf("%s://%s", issuerURL.Scheme, issuerURL.Hostname())
	if issuerURL.Port() != "" {
		finalUrl = fmt.Sprintf("%s:%s", finalUrl, issuerURL.Port())
	}

	return fmt.Sprintf("%s/dex", finalUrl), nil
}
