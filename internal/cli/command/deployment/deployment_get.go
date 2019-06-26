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

type getDeploymentOptions struct {
	deploymentOptions

	deploymentReleaseName string
}

// NewDeploymentGetCommand returns a `*cobra.Command` for `banzai deployment get` subcommand.
func NewDeploymentGetCommand(banzaiCli cli.Cli) *cobra.Command {
	options := getDeploymentOptions{}

	cmd := &cobra.Command{
		Use:           "get RELEASE-NAME",
		Short:         "Get deployment details",
		Long:          "Get the details of a deployment identified by deployment release name. In order to display deployment current values and notes use --output=(json|yaml)",
		Args:          cobra.ExactArgs(1),
		Aliases:       []string{"g", "show"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")
			options.deploymentReleaseName = args[0]

			return runGetDeployment(banzaiCli, options)
		},
		Example: `
			$ banzai deployment get dns
			? Cluster  [Use arrows to move, type to filter]
			> pke-cluster-1
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2 

			$ banzai deployment get dns --cluster-name pke-cluster-1
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2

			$ banzai deployment get dns --cluster 1846
			
			Namespace        ReleaseName  Status    Version  UpdatedAt             CreatedAt             ChartName     ChartVersion
			pipeline-system  dns          DEPLOYED  1        2019-06-23T06:52:24Z  2019-06-23T06:52:24Z  external-dns  1.6.2 
`,
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.clusterName, "cluster-name", "n", "", "Name of the cluster to get deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster to get deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")

	return cmd
}

func runGetDeployment(banzaiCli cli.Cli, options getDeploymentOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	clusterID, err := getClusterID(banzaiCli, orgID, options.deploymentOptions)
	if err != nil {
		return err
	}

	deployment, _, err := banzaiCli.Client().DeploymentsApi.GetDeployment(context.Background(), orgID, clusterID,options.deploymentReleaseName, nil)
	if err != nil {
		return emperror.Wrap(err, "could not get deployment details")
	}

	err = format.DeploymentWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), deployment)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out deployment")
	}

	return nil
}
