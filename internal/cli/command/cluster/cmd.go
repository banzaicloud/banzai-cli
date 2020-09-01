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

package cluster

import (
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/node"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/nodepool"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/restore"
)

// NewClusterCommand returns a cobra command for `cluster` subcommands.
func NewClusterCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters", "c", "cl"},
		Short:   "Manage clusters",
	}

	cmd.AddCommand(
		NewCreateCommand(banzaiCli),
		NewDeleteCommand(banzaiCli),
		NewGetCommand(banzaiCli),
		NewUpdateCommand(banzaiCli),
		NewHelmCommand(banzaiCli),
		NewImportCommand(banzaiCli),
		NewListCommand(banzaiCli),
		NewShellCommand(banzaiCli),
		NewConfigCommand(banzaiCli),
		integratedservice.NewIntegratedServiceCommand(banzaiCli),
		node.NewNodeCommand(banzaiCli),
		nodepool.NewNodePoolCommand(banzaiCli),
		restore.NewRestoreCommand(banzaiCli),
	)

	return cmd
}
