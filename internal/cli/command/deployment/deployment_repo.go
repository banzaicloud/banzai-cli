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

package deployment

import (
	"context"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/spf13/cobra"
)

type listDeploymentRepoOptions struct {
	deploymentOptions
}

// NewDeploymentListCommand returns a `*cobra.Command` for `banzai deployment list` subcommand.
func NewDeploymentRepoCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listDeploymentRepoOptions{}

	cmd := &cobra.Command{
		Use:     "repo",
		Short:   "List repos",
		Args:    cobra.NoArgs,
		Aliases: []string{"rp"},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")

			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return runListDeploymentsRepos(banzaiCli, options)
		},
		Example: `
				$ banzai deployment repo`,
	}

	flags := cmd.Flags()

	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster which to list deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")

	return cmd
}

func runListDeploymentsRepos(banzaiCli cli.Cli, options listDeploymentRepoOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	repos, _, err := banzaiCli.Client().HelmApi.HelmListRepos(context.Background(), orgID)
	if err != nil {
		return emperror.Wrap(err, "could not list helm repositories")
	}

	err = format.HelmReposWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), repos)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out helm repositories")
	}

	return nil
}
