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

package ingress

import (
	"context"
	"strings"

	"emperror.dev/errors"

	"github.com/banzaicloud/banzai-cli/internal/cli"
)

func splitCommaSeparatedList(s string) []string {
	if s == "" {
		return nil
	}

	list := strings.Split(s, ",")
	for i, e := range list {
		list[i] = strings.TrimSpace(e)
	}
	return list
}

func getAvailableControllerTypes(ctx context.Context, banzaiCLI cli.Cli) ([]string, error) {
	caps, _, err := banzaiCLI.Client().PipelineApi.ListCapabilities(ctx)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to list capabilities")
	}

	featuresCaps, ok := caps["features"]
	if !ok {
		return nil, errors.New("failed to get features' capabilities")
	}

	ingressCaps, ok := featuresCaps["ingress"].(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to get ingress capabilities")
	}

	controllers, ok := ingressCaps["controllers"].([]interface{})
	if !ok {
		return nil, errors.New("failed to get list of available ingress controllers")
	}

	result := make([]string, len(controllers))
	for i, c := range controllers {
		s, ok := c.(string)
		if !ok {
			return nil, errors.Errorf("not a valid ingress controller type: %v", c)
		}
		result[i] = s
	}

	return result, nil
}
