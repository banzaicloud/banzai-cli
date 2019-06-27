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
)

type createDeploymentOptions struct {
	deploymentOptions

	file string
}

// NewDeploymentCreateCommand returns a `*cobra.Command` for `banzai deployment create` subcommand.
func NewDeploymentCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createDeploymentOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Creates a deployment",
		Long:    "Creates a deployment based on deployment descriptor JSON read from stdin or file.",
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")

			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return runCreateDeployment(banzaiCli, options)
		},
		Example: `
        # Create deployment from file using interactive mode
        ----------------------------------------------------
        $ banzai deployment create
        ? Cluster  [Use arrows to move, type to filter]
        > pke-cluster-1
        ? Load a JSON or YAML file: [? for help] /var/tmp/wordpress.json

        ReleaseName       Notes
        torpid-armadillo  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK


        # Create deployment from stdin
        ------------------------------
        $ banzai deployment create --cluster-name pke-cluster-1 -f -<<EOF
        > {
        >   "name": "stable/wordpress",
        >   "releasename": "",
        >   "namespace": "default",
        >   "version": "5.12.4",
        >   "dryRun": false,
        >   "values": {
        >		"replicaCount": 2
        >   }
        > }
        > EOF

        ReleaseName       Notes
        lumbering-lizard  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK  

        $ echo '{"name":"stable/wordpress","releasename":"my-wordpress-1"}' |  banzai deployment create --cluster-name pke-cluster-1
        ReleaseName     Notes
        my-wordpress-1  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK

        # Create deployment from file
        -----------------------------
        $ banzai deployment create --cluster-name pke-cluster-1 --file /var/tmp/wordpress.json --no-interactive

        ReleaseName         Notes
        eyewitness-opossum  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK`,
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.clusterName, "cluster-name", "n", "", "Name of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.StringVarP(&options.file, "file", "f", "", "Deployment descriptor file")

	return cmd
}

func runCreateDeployment(banzaiCli cli.Cli, options createDeploymentOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	clusterID, err := getClusterID(banzaiCli, orgID, options.deploymentOptions)
	if err != nil {
		return err
	}

	req, err := buildCreateUpdateDeploymentRequest(banzaiCli, options.file)
	if err != nil {
		return emperror.Wrap(err, "could not prepare deployment creation request")
	}

	if req == nil {
		return errors.New("missing deployment descriptor")
	}

	response, _, err := banzaiCli.Client().DeploymentsApi.CreateDeployment(context.Background(), orgID, clusterID, *req)
	if err != nil {
		return emperror.Wrap(err, "could not create deployment")
	}

	err = format.DeploymentCreateUpdateResponseWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), response)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out deployment create response")
	}

	return nil
}
