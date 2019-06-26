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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type deleteDeploymentOptions struct {
	deploymentOptions

	releaseName string
}

// NewDeploymentDeleteCommand returns a `*cobra.Command` for `banzai cluster deployment delete` subcommand.
func NewDeploymentDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deleteDeploymentOptions{}

	cmd := &cobra.Command{
		Use:           "delete RELEASE-NAME",
		Short:         "Delete a deployment",
		Long:          "Delete a deployment identified by deployment release name.",
		Args:          cobra.ExactArgs(1),
		Aliases:       []string{"del", "rm"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")
			options.releaseName = args[0]

			return runDeleteDeployment(banzaiCli, options)
		},
		Example: `
			$ banzai deployment delete test-deployment
			? Cluster  [Use arrows to move, type to filter]
			> pke-cluster-1

			Name  			 Status  Message            
			test-deployment  200     Deployment deleted!

			$ banzai deployment delete test-deployment --cluster-name pke-cluster-1
			Name  			 Status  Message            
			test-deployment  200     Deployment deleted!`,
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.clusterName, "cluster-name", "n", "", "Name of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")

	return cmd
}

func runDeleteDeployment(banzaiCli cli.Cli, options deleteDeploymentOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	clusterID, err := getClusterID(banzaiCli, orgID, options.deploymentOptions)
	if err != nil {
		return err
	}

	releaseName, err := getReleaseName(banzaiCli, orgID, clusterID, options.releaseName)
	if err != nil {
		return err
	}

	confirmed := false
	survey.AskOne(&survey.Confirm{Message: "Do you want to DELETE the deployment?"}, &confirmed, nil)
	if !confirmed {
		return errors.New("deletion cancelled")
	}

	deployment, _, err := banzaiCli.Client().DeploymentsApi.DeleteDeployment(context.Background(), orgID, clusterID, releaseName)
	if err != nil {
		return emperror.Wrap(err, "could not delete deployment")
	}

	err = format.DeploymentDeleteResponseWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), deployment)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out deployment deletion status")
	}

	return nil
}
