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

package node

import (
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

var NodeClusterContext clustercontext.Context

// NewNodeCommand returns a cobra command for `node` subcommands.
func NewNodeCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"nodes", "n"},
		Short:   "Work with cluster nodes",
	}

	cmd.AddCommand(
		NewNodeListCommand(banzaiCli),
		NewSSHToNodeCommand(banzaiCli),
	)

	NodeClusterContext = clustercontext.NewClusterContext(cmd, banzaiCli, "node")

	return cmd
}
