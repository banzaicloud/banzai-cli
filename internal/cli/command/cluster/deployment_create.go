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
	"io/ioutil"
	"os"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type createDeploymentOptions struct {
	deploymentOptions

	file string
}

// NewDeploymentCreateCommand returns a `*cobra.Command` for `banzai cluster deployment create` subcommand.
func NewDeploymentCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createDeploymentOptions{}

	cmd := &cobra.Command{
		Use:           "create",
		Short:         "Creates a deployment",
		Long:          "Creates a deployment based on deployment descriptor JSON read from stdin or file.",
		Args:          cobra.MaximumNArgs(1),
		Aliases:       []string{"c"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")

			return runCreateDeployment(banzaiCli, options)
		},
		Example: `
        # Create deployment from file using interactive mode
        ----------------------------------------------------
        $ banzai cluster deployment create
        ? Cluster  [Use arrows to move, type to filter]
        > pke-cluster-1
        ? Load a JSON or YAML file: [? for help] /var/tmp/wordpress.json

        ReleaseName       Notes
        torpid-armadillo  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK


        # Create deployment from stdin
        ------------------------------
        $ banzai cluster deployment create --cluster-name pke-cluster-1 -f -<<EOF
        > {
        >   "name": "stable/wordpress",
        >   "releasename": "",
        >   "namespace": "default",
        >   "version": "5.12.4",
        >   "dryRun": false
        > }
        > EOF

        ReleaseName       Notes
        lumbering-lizard  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK  

        $ echo '{"name":"stable/wordpress","releasename":"my-wordpress-1"}' |  banzai cluster deployment create --cluster-name pke-cluster-1
        ReleaseName     Notes
        my-wordpress-1  V29yZHByZXNzIGRlcGxveW1lbnQgbm90ZXMK

        # Create deployment from file using non interactive mode
        --------------------------------------------------------
        $ build/banzai cluster deployment create --cluster-name pke-cluster-1 --file /var/tmp/wordpress.json --no-interactive

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

	req, err := getCreateDeploymentRequest(banzaiCli, options)
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
		return emperror.Wrap(err, "cloud not print out deployment creation status")
	}

	return nil
}

func getCreateDeploymentRequest(banzaiCli cli.Cli, options createDeploymentOptions) (*pipeline.CreateUpdateDeploymentRequest, error) {
	var raw []byte
	var err error
	var fileName = options.file

	if banzaiCli.Interactive() {
		for {
			if options.file == "" {
				_ = survey.AskOne(
					&survey.Input{
						Message: "Load a JSON or YAML file:",
						Default: "",
						Help:    "Give either a relative or an absolute path to a file containing a JSON or YAML deployment request.",
					},
					&fileName,
					nil,
				)
				if fileName == "" {
					return nil, nil
				}

				raw, err = ioutil.ReadFile(fileName)
				if err != nil {
					fileName = "" // reset fileName so that we can ask for one

					log.Errorf("failed to read file %q: %v", fileName, err)

					continue
				} else {
					break
				}
			}
		}
	} else {
		if fileName != "" && fileName != "-" {
			raw, err = ioutil.ReadFile(fileName)
		} else {
			raw, err = ioutil.ReadAll(os.Stdin)
			fileName = "stdin"
		}

		if err != nil {
			return nil, emperror.Wrapf(err, "failed to read file %q", fileName)
		}

	}

	req, err := unmarshalCreateUpdateDeploymentRequest(raw)
	if err != nil {
		return nil, emperror.Wrap(err, "could not parse deployment request")
	}

	return req, nil


}
