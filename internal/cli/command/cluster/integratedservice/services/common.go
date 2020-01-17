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

package services

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/utils"
)

const (
	serviceKeyOnCap = "features"
	enabledKeyOnCap = "enabled"
)

type Cap map[string]interface{}

type capLoader struct {
	cli cli.Cli
}

func (capabilities Cap) isServiceEnabled() error {
	if en, ok := capabilities[enabledKeyOnCap]; ok {
		if enabled, ok := en.(bool); ok {
			if enabled {
				return nil
			}
		}
	}

	return errors.New("service disabled")
}

func (cl capLoader) loadCapabilities(ctx context.Context, serviceName string) (Cap, error) {
	capabilities, r, err := cl.cli.Client().PipelineApi.ListCapabilities(ctx)
	if err := utils.CheckCallResults(r, err); err != nil {
		return nil, errors.WrapIf(err, "failed to retrieve capabilities")
	}

	if c, ok := capabilities[serviceKeyOnCap]; ok {
		if s, ok := c[serviceName]; ok {
			if svc, ok := s.(map[string]interface{}); ok {
				return svc, nil
			}
		}
	}

	return nil, errors.New(fmt.Sprintf("service %q disabled", serviceName))
}
