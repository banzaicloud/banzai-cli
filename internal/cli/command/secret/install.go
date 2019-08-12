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

package secret

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type installSecretOptions struct {
	file       string
	secretName string
	merge      bool
	clustercontext.Context
}

func NewInstallCommand(banzaiCli cli.Cli) *cobra.Command {
	options := installSecretOptions{}

	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "Install a secret to a cluster",
		Long:    "Install a particular secret from Pipeline as a Kubernetes secret to a cluster.",
		Example: `
		Install secret
		-----
		$ banzai secret install --name mysecretname --cluster-name myClusterName <<EOF
		> {
		> 	"namespace": "default",
		> 	"spec": {
		> 		"ROOT_USER": {
		> 			"source": "AWS_ACCESS_KEY_ID"
		> 		}
		> 	}
		> }
		> EOF
		`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return runInstallSecret(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Template descriptor file")
	flags.StringVarP(&options.secretName, "name", "n", "", "Name of the Pipeline secret to use")
	flags.BoolVarP(&options.merge, "merge", "m", false, "Merge fields to an existing Kubernetes secret")
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "install secret on")

	return cmd
}

func runInstallSecret(banzaiCli cli.Cli, options installSecretOptions) error {

	if options.secretName == "" {
		return errors.New("secret name is not specified (--name)") // TODO: ask from existing secrets
	}

	if err := options.Init(); err != nil {
		return errors.WrapIf(err, "failed to select cluster")
	}

	out := &pipeline.InstallSecretRequest{
		Namespace:        "default",
		SourceSecretName: options.secretName, // set defaults explicitly
	}

	if banzaiCli.Interactive() {
		err := buildInteractiveInstallSecretRequest(options, out)
		if err != nil {
			return err
		}
	} else {
		// non-interactive
		filename, raw, err := utils.ReadFileOrStdin(options.file)
		if err != nil {
			return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
		}

		log.Debugf("%d bytes read", len(raw))

		if err := validateInstallSecretRequest(raw); err != nil {
			return errors.WrapIf(err, "failed to parse create cluster request")
		}

		if err := utils.Unmarshal(raw, &out); err != nil {
			return errors.WrapIf(err, "failed to unmarshal create cluster request")
		}

	}

	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()
	log.Debugf("sending install secret request: %#v", out)

	_, response, err := banzaiCli.Client().ClustersApi.InstallSecret(context.Background(), orgID, clusterID, options.secretName, *out)
	if response != nil && response.StatusCode == http.StatusConflict {
		log.Infof("Secret (%s) already installed to cluster (%s)", options.secretName, options.ClusterName())

		if options.merge {
			if _, _, err = banzaiCli.Client().ClustersApi.MergeSecret(context.Background(), orgID, clusterID, options.secretName, *out); err != nil {
				cli.LogAPIError("merge secret", err, out)
				return errors.WrapIf(err, "failed to merge secret")
			}
		} else {
			return errors.New("set --merge flag to merge existing secret")
		}
	}

	if err != nil {
		cli.LogAPIError("install secret", err, out)
		return errors.WrapIf(err, "failed to install secret")
	}

	log.Info("secret installed to cluster")

	return nil
}

func buildInteractiveInstallSecretRequest(options installSecretOptions, out *pipeline.InstallSecretRequest) error {
	var fileName = options.file

	for {
		if fileName == "" {
			_ = survey.AskOne(
				&survey.Input{
					Message: "Load a JSON or YAML file:",
					Default: "skip",
					Help:    "Give either a relative or an absolute path to a file containing a JSON or YAML secret installation request. Leave empty to cancel.",
				},
				&fileName,
			)
			if fileName == "skip" || fileName == "" {
				break
			}
		}

		if raw, err := ioutil.ReadFile(fileName); err != nil {
			fileName = "" // reset fileName so that we can ask for one

			log.Errorf("failed to read file %q: %v", fileName, err)

			continue
		} else {
			if err := utils.Unmarshal(raw, out); err != nil {
				return errors.WrapIf(err, "failed to parse InstallSecretRequest")
			}

			break
		}
	}

	return nil
}

func validateInstallSecretRequest(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		if bytes, ok := val.([]byte); ok {
			str = string(bytes)
		} else {
			return errors.New("value is not a string or []byte")
		}
	}

	decoder := json.NewDecoder(strings.NewReader(str))

	var typer struct{ Type string }
	err := decoder.Decode(&typer)
	if err != nil {
		return errors.WrapIf(err, "invalid JSON request")
	}

	decoder = json.NewDecoder(strings.NewReader(str))
	decoder.DisallowUnknownFields()

	return errors.WrapIf(decoder.Decode(&pipeline.InstallSecretRequest{}), "invalid request")
}
