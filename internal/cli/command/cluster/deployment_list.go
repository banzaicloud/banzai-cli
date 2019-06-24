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


type listDeploymentOptions struct {
	deploymentOptions
}

// NewDeploymentListCommand returns a `*cobra.Command` for `banzai cluster deployment list` subcommand.
func NewDeploymentListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listDeploymentOptions{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List deployments",
		Args:          cobra.NoArgs,
		Aliases:       []string{"l", "ls"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")

			return runListDeployments(banzaiCli, options)
		},
		Example: `
				$ banzai cluster deployment ls

				? Cluster  [Use arrows to move, type to filter]
				> pke-cluster-1

				Namespace        ReleaseName     Status    Version  UpdatedAt             CreatedAt             ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        2019-06-23T06:53:00Z  2019-06-23T06:53:00Z  anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        2019-06-23T06:52:57Z  2019-06-23T06:52:57Z  pipeline-cluster-monitor      0.1.17      
				pipeline-system  hpa-operator    DEPLOYED  1        2019-06-23T06:52:29Z  2019-06-23T06:52:29Z  hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        2019-06-23T06:52:28Z  2019-06-23T06:52:28Z  cluster-autoscaler            0.12.3      

				$ banzai cluster deployment ls --cluster-name pke-cluster-1

				Namespace        ReleaseName     Status    Version  UpdatedAt             CreatedAt             ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        2019-06-23T06:53:00Z  2019-06-23T06:53:00Z  anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        2019-06-23T06:52:57Z  2019-06-23T06:52:57Z  pipeline-cluster-monitor      0.1.17      
				pipeline-system  hpa-operator    DEPLOYED  1        2019-06-23T06:52:29Z  2019-06-23T06:52:29Z  hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        2019-06-23T06:52:28Z  2019-06-23T06:52:28Z  cluster-autoscaler            0.12.3

				$ banzai cluster deployment ls --cluster 1846

				Namespace        ReleaseName     Status    Version  UpdatedAt             CreatedAt             ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        2019-06-23T06:53:00Z  2019-06-23T06:53:00Z  anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        2019-06-23T06:52:57Z  2019-06-23T06:52:57Z  pipeline-cluster-monitor      0.1.17      
				pipeline-system  hpa-operator    DEPLOYED  1        2019-06-23T06:52:29Z  2019-06-23T06:52:29Z  hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        2019-06-23T06:52:28Z  2019-06-23T06:52:28Z  cluster-autoscaler            0.12.3`,
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.clusterName, "cluster-name", "n", "", "Name of the cluster which to list deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster which to list deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")

	return cmd
}

func runListDeployments(banzaiCli cli.Cli, options listDeploymentOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	clusterID, err := getClusterID(banzaiCli, orgID, options.deploymentOptions)
	if err != nil {
		return err
	}


	deployments, _, err := banzaiCli.Client().DeploymentsApi.ListDeployments(context.Background(), orgID, clusterID, nil)
	if err != nil {
		return emperror.Wrap(err, "could not list deployments")
	}

	err = format.DeploymentsWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), deployments)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out deployments")
	}

	return nil
}
