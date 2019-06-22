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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)


type listOptions struct {
	clusterName string
	clusterID int32

	format string
}

// NewDeploymentListCommand returns a `*cobra.Command` for `banzai cluster deployment list` subcommand.
func NewDeploymentListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List deployments",
		Args:          cobra.NoArgs,
		Aliases:       []string{"l", "ls"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")

			return runList(banzaiCli, options)
		},
		Example: `
				$ banzai cluster deployment ls

				? Cluster  [Use arrows to move, type to filter]
				> pke-cluster-1

				Namespace        ReleaseName     Status    Version  ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        pipeline-cluster-monitor      0.1.17
				pipeline-system  hpa-operator    DEPLOYED  1        hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        cluster-autoscaler            0.12.3      

				$ banzai cluster deployment ls --cluster-name pke-cluster-1 --no-interactive

				Namespace        ReleaseName     Status    Version  ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        pipeline-cluster-monitor      0.1.17
				pipeline-system  hpa-operator    DEPLOYED  1        hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        cluster-autoscaler            0.12.3

				$ banzai cluster deployment ls --cluster 1846 --no-interactive

				Namespace        ReleaseName     Status    Version  ChartName                     ChartVersion
				pipeline-system  anchore         DEPLOYED  1        anchore-policy-validator      0.3.5       
				pipeline-system  monitor         DEPLOYED  1        pipeline-cluster-monitor      0.1.17
				pipeline-system  hpa-operator    DEPLOYED  1        hpa-operator                  0.0.10      
				kube-system      autoscaler      DEPLOYED  1        cluster-autoscaler            0.12.3`,
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.clusterName, "cluster-name", "n", "", "Name of the cluster to list deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster to list deployments from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")

	return cmd
}

func runList(banzaiCli cli.Cli, options listOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	var clusterID int32
	var err error

	if banzaiCli.Interactive() {
		clusterID, err = input.AskCluster(banzaiCli, orgID, options.clusterID, options.clusterName)
		if err != nil {
			return emperror.Wrap(err, "could not ask for a cluster")
		}
	} else if options.clusterID > 0 {
		// check if cluster exists
		_, err = banzaiCli.Client().ClustersApi.GetClusterStatus(context.Background(), orgID, options.clusterID)
		if err == nil {
			clusterID = options.clusterID
		}
	} else if options.clusterName != "" {
		// check if cluster exists
		clusters, _, err := banzaiCli.Client().ClustersApi.ListClusters(context.Background(), orgID)
		if err != nil {
			cli.LogAPIError("list clusters", err, orgID)
			return emperror.Wrap(err, "could not list clusters")
		}

		for _, cluster := range clusters {
			if cluster.Name == options.clusterName {
				clusterID = cluster.Id
				break
			}
		}
	} else {
		return errors.New("No cluster is specified. Use the --cluster or --cluster-name option or run the CLI in interactive mode")
	}

	if clusterID == 0 {
		return errors.New("cluster could not be found")
	}


	deployments, _, err := banzaiCli.Client().DeploymentsApi.ListDeployments(context.Background(), orgID, clusterID, nil)
	if err != nil {
		return emperror.Wrap(err, "could not list deployments")
	}

	// received response contains camel case field names

	err = format.DeploymentsWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), deployments)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out deployments")
	}

	return nil
}
