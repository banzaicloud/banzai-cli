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
	"context"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/spf13/cobra"
)

type getOptions struct {
	deploymentOptions

	deploymentName string

}

// NewDeploymentGetCommand returns a `*cobra.Command` for `banzai cluster deployment get` subcommand.
func NewDeploymentGetCommand(banzaiCli cli.Cli) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:           "get NAME",
		Short:         "Get deployment details",
		Long:          "In order to display deployment current values and notes use --output=(json|yaml)",
		Args:          cobra.ExactArgs(1),
		Aliases:       []string{"g", "show"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")
			options.deploymentName = args[0]

			return runGet(banzaiCli, options)
		},
		Example: `
			$ banzai cluster deployment get dns
			? Cluster  [Use arrows to move, type to filter]
			> pke-cluster-1
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2 

			$ banzai cluster deployment get dns --cluster-name pke-cluster-1 --no-interactive
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2

			$ banzai cluster deployment get dns --cluster 1846 --no-interactive
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2 
`,
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.clusterName, "cluster-name", "n", "", "Name of the cluster to get deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster to get deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")

	return cmd
}

func runGet(banzaiCli cli.Cli, options getOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	clusterID, err := getClusterID(banzaiCli, orgID, options.deploymentOptions)
	if err != nil {
		return err
	}

	deployment, _, err := banzaiCli.Client().DeploymentsApi.GetDeployment(context.Background(), orgID, clusterID,options.deploymentName, nil)
	if err != nil {
		return emperror.Wrap(err, "could not get deployment details")
	}

	err = format.DeploymentWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), deployment)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out deployment")
	}

	return nil
}
