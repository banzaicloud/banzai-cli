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
	"net/url"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/auth"
)

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

	lApp := auth.NewLoginApp(banzaiCli, issuerURL.String(), pipelineBasePath)
	tokenBytes, err := auth.RunAuthServer(lApp)
	if err != nil {
		return "", errors.WrapIf(err, "login failed")
	}

	return string(tokenBytes), nil
}
