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
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/utils"
)

const (
	serviceKeyOnCap = "features"
	enabledKeyOnCap = "enabled"
)

type ServiceCommandManager interface {
	BuildActivateRequestInteractively(banzaiCli cli.Cli, clusterCtx clustercontext.Context) (pipeline.ActivateIntegratedServiceRequest, error)
	BuildUpdateRequestInteractively(banzaiCli cli.Cli, updateServiceRequest *pipeline.UpdateIntegratedServiceRequest, clusterCtx clustercontext.Context) error
	ReadableName() string
	ServiceName() string
	WriteDetailsTable(details pipeline.IntegratedServiceDetails) map[string]map[string]interface{}
	specValidator
}

func NewServiceCommand(banzaiCLI cli.Cli, use string, scm ServiceCommandManager) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Manage cluster %s service", scm.ReadableName()),
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runGet(banzaiCLI, scm, options, args, use)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCLI, fmt.Sprintf("manage %s cluster service of", scm.ReadableName()))

	cmd.AddCommand(
		newGetCommand(banzaiCLI, use, scm),
		newActivateCommand(banzaiCLI, use, scm),
		newDeactivateCommand(banzaiCLI, use, scm),
		newUpdateCommand(banzaiCLI, use, scm),
	)

	return cmd
}

type specValidator interface {
	ValidateSpec(spec map[string]interface{}) error
}

func isServiceEnabled(ctx context.Context, banzaiCLI cli.Cli, serviceName string) error {
	capabilities, r, err := banzaiCLI.Client().PipelineApi.ListCapabilities(ctx)
	if err := utils.CheckCallResults(r, err); err != nil {
		return errors.WrapIf(err, "failed to retrieve capabilities")
	}

	if services, ok := capabilities[serviceKeyOnCap]; ok {
		if s, ok := services[serviceName]; ok {
			if svc, ok := s.(map[string]interface{}); ok {
				if en, ok := svc[enabledKeyOnCap]; ok {
					if enabled, ok := en.(bool); ok {
						if enabled {
							return nil
						}
					}
				}
			}
		}
	}

	return errors.New(fmt.Sprintf("%s service disabled", serviceName))
}
