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

package feature

import (
	"fmt"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/feature/features"
	"github.com/spf13/cobra"
)

func NewFeatureCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:     "feature",
		Aliases: []string{"features", "feat", "ft"},
		Short:   "Manage cluster features",
		Args:    cobra.MaximumNArgs(1),
		Hidden:  true,
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "list features")

	cmd.AddCommand(
		NewListCommand(banzaiCli),
		// NOTE: add feature commands here
		featureCommandFactory(banzaiCli, "dns", features.NewDNSSubCommandManager()),
		featureCommandFactory(banzaiCli, "vault", features.NewVaultSubCommandManager()),
		featureCommandFactory(banzaiCli, "monitoring", features.NewMonitoringSubCommandManager()),
	)

	return cmd
}

type getOptions struct {
	clustercontext.Context
}

type SubCommandManager interface {
	GetName() string
	ActivateManager() features.ActivateManager
	DeactivateManager() features.DeactivateManager
	GetManager() features.GetManager
	UpdateManager() features.UpdateManager
}

func featureCommandFactory(banzaiCLI cli.Cli, use string, scm SubCommandManager) *cobra.Command {
	options := getOptions{}
	getCommand := features.GetCommandFactory(banzaiCLI, scm.GetManager(), scm.GetName())

	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Manage cluster %s feature", scm.GetName()),
		Args:  cobra.NoArgs,
		RunE:  getCommand.RunE,
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCLI, fmt.Sprintf("manage %s cluster feature of", scm.GetName()))

	cmd.AddCommand(
		features.ActivateCommandFactory(banzaiCLI, scm.ActivateManager(), scm.GetName()),
		features.DeactivateCommandFactory(banzaiCLI, scm.DeactivateManager(), scm.GetName()),
		getCommand,
		features.UpdateCommandFactory(banzaiCLI, scm.UpdateManager(), scm.GetName()),
	)

	return cmd
}
