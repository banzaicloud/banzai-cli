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

type updateDeploymentOptions struct {
	deploymentOptions

	file        string
	releaseName string
}

// NewDeploymentUpdateCommand returns a `*cobra.Command` for `banzai deployment create` subcommand.
func NewDeploymentUpdateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := updateDeploymentOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Updates a deployment",
		Long:    "Updates a deployment identified by release name using a deployment descriptor JSON read from stdin or file.",
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")

			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return runUpdateDeployment(banzaiCli, options)
		},
		Example: `
		# Update deployment from file using interactive mode
        ----------------------------------------------------
        $ banzai deployment update
        ? Cluster pke-cluster-1
        ? Release name  [Use arrows to move, type to filter]
        > hazelcast-1
        exacerbated-narwhal
        luminous-hare

        ? Load a JSON or YAML file: [? for help] /var/tmp/hazelcast.json

        ReleaseName  Notes
        hazelcast-1  aGF6ZWxjYXN0LTEgcmVsZWFzZQo=

        # Update deployment from stdin
        ------------------------------
        $ banzai deployment update --cluster-name pke-cluster-1 --release-name hazelcast-1 -f -<<EOF
        > {
        >     "name": "stable/hazelcast",
        >     "version": "1.3.3",
        >     "reuseValues": true,
        >     "values": {
        >         "cluster": {
        >             "memberCount": 5
        >         }
        >     } 
        > }
        > EOF

        $ echo '{"name":"stable/hazelcast","version":"1.3.3","reuseValues":true,"values":{"cluster":{"memberCount":5}}}' | banzai deployment update --cluster-name pke-cluster-1 --release-name hazelcast-1

        # Update deployment from file
        -----------------------------
        $ banzai deployment update --cluster-name pke-cluster-1 --release-name hazelcast-1 -f /var/tmp/hazelcast.json`,
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.clusterName, "cluster-name", "n", "", "Name of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.Int32VarP(&options.clusterID, "cluster", "", 0, "ID of the cluster to delete deployment from. Specify either --cluster-name or --cluster. In interactive mode the CLI prompts the user to select a cluster")
	flags.StringVarP(&options.file, "file", "f", "", "Deployment descriptor file")
	flags.StringVarP(&options.releaseName, "release-name", "r", "", "Deployment release name")

	return cmd
}

func runUpdateDeployment(banzaiCli cli.Cli, options updateDeploymentOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	clusterID, err := getClusterID(banzaiCli, orgID, options.deploymentOptions)
	if err != nil {
		return err
	}

	releaseName, err := getReleaseName(banzaiCli, orgID, clusterID, options.releaseName)
	if err != nil {
		return err
	}

	req, err := buildCreateUpdateDeploymentRequest(banzaiCli, options.file)
	if err != nil {
		return emperror.Wrap(err, "could not prepare deployment update request")
	}

	if req == nil {
		return errors.New("missing deployment descriptor")
	}

	response, _, err := banzaiCli.Client().DeploymentsApi.UpdateDeployment(context.Background(), orgID, clusterID, releaseName, *req)
	if err != nil {
		return emperror.Wrap(err, "could not update deployment")
	}

	err = format.DeploymentCreateUpdateResponseWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), response)
	if err != nil {
		return emperror.Wrap(err, "cloud not print out deployment update response")
	}

	return nil
}



