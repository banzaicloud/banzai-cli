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

package integratedservice

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services"
)

func NewIntegratedServiceCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc", "integratedservice", "is"},
		Short:   "Manage cluster integrated services",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "list services")

	cmd.AddCommand(
		NewListCommand(banzaiCli),
		// NOTE: add integratedservice commands here
		serviceCommandFactory(banzaiCli, "dns", services.NewDNSSubCommandManager()),
		serviceCommandFactory(banzaiCli, "vault", services.NewVaultSubCommandManager()),
		serviceCommandFactory(banzaiCli, "securityscan", services.NewSecurityScanSubCommandManager()),
		serviceCommandFactory(banzaiCli, "monitoring", services.NewMonitoringSubCommandManager()),
		serviceCommandFactory(banzaiCli, "logging", services.NewLoggingSubCommandManager()),
		serviceCommandFactory(banzaiCli, "expiry", services.NewExpirySubCommandManager()),
	)

	return cmd
}

type getOptions struct {
	clustercontext.Context
}

type SubCommandManager interface {
	GetName() string
	ActivateManager() services.ActivateManager
	DeactivateManager() services.DeactivateManager
	GetManager() services.GetManager
	UpdateManager() services.UpdateManager
}

func serviceCommandFactory(banzaiCLI cli.Cli, use string, scm SubCommandManager) *cobra.Command {
	options := services.GetOptions{}

	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Manage cluster %s service", scm.GetName()),
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return services.RunGet(banzaiCLI, scm.GetManager(), options, args, use)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCLI, fmt.Sprintf("manage %s cluster service of", scm.GetName()))

	cmd.AddCommand(
		services.GetCommandFactory(banzaiCLI, use, scm.GetManager(), scm.GetName()),
		services.ActivateCommandFactory(banzaiCLI, use, scm.ActivateManager(), scm.GetName()),
		services.DeactivateCommandFactory(banzaiCLI, use, scm.DeactivateManager(), scm.GetName()),
		services.UpdateCommandFactory(banzaiCLI, use, scm.UpdateManager(), scm.GetName()),
	)

	return cmd
}
