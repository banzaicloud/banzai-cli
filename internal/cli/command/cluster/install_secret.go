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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type installSecretOptions struct {
	file        string
	secretName  string
	clusterName string
	merge       bool
}

func NewSecretInstallCommand(banzaiCli cli.Cli) *cobra.Command {
	options := installSecretOptions{}

	cmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"s"},
		Short:   "Install a secret to a cluster",
		Long:    "Install a particular secret to a cluster's namespace.",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstallSecret(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Template descriptor file")
	flags.StringVarP(&options.secretName, "secret-name", "s", "", "Name of the secret to install")
	flags.StringVarP(&options.clusterName, "cluster-name", "c", "", "Name of the cluster to install the secret")
	flags.BoolVarP(&options.merge, "merge", "m", false, "Set true to merge existing secret")

	return cmd
}

func runInstallSecret(banzaiCli cli.Cli, options installSecretOptions) error {
	out := &pipeline.InstallSecretRequest{}

	if banzaiCli.Interactive() {
		err := buildInteractiveInstallSecretRequest(options, out)
		if err != nil {
			return err
		}
	} else {
		// non-interactive
		var raw []byte
		var err error
		filename := options.file

		if filename != "" && filename != "-" {
			raw, err = ioutil.ReadFile(filename)
		} else {
			raw, err = ioutil.ReadAll(os.Stdin)
			filename = "stdin"
		}

		log.Debugf("%d bytes read", len(raw))

		if err != nil {
			return emperror.WrapWith(err, fmt.Sprintf("failed to read %q", filename), "filename", filename)
		}

		if err := validateInstallSecretRequest(raw); err != nil {
			return emperror.Wrap(err, "failed to parse create cluster request")
		}

		if err := unmarshal(raw, &out); err != nil {
			return emperror.Wrap(err, "failed to unmarshal create cluster request")
		}

	}

	log.Debugf("install secret request: %#v", out)

	// find cluster
	orgID := input.GetOrganization(banzaiCli)
	clusters, _, err := banzaiCli.Client().ClustersApi.ListClusters(context.Background(), orgID)
	if err != nil {
		cli.LogAPIError("list clusters", err, orgID)
		log.Fatalf("could not list clusters: %v", err)
	}
	var clusterId int32
	for _, cluster := range clusters {
		if cluster.Name == options.clusterName {
			clusterId = cluster.Id
			break
		}
	}

	_, response, err := banzaiCli.Client().ClustersApi.InstallSecret(context.Background(), orgID, clusterId, options.secretName, *out)
	if response != nil && response.StatusCode == http.StatusConflict {
		log.Infof("Secret (%s) already installed to cluster (%s)", options.secretName, options.clusterName)

		if options.merge {
			log.Info("path secret")
			if _, _, err = banzaiCli.Client().ClustersApi.MergeSecret(context.Background(), orgID, clusterId, options.secretName, *out); err != nil {
				cli.LogAPIError("merge secret", err, out)
				return emperror.Wrap(err, "failed to merge secret")
			}
		} else {
			return errors.New("set -merge flag to true to merge existing secret")
		}
	}

	if err != nil {
		cli.LogAPIError("install secret", err, out)
		return emperror.Wrap(err, "failed to install secret")
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
				nil,
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
			if err := unmarshal(raw, &out); err != nil {
				return emperror.Wrap(err, "failed to parse InstallSecretRequest")
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
		return emperror.Wrap(err, "invalid JSON request")
	}

	decoder = json.NewDecoder(strings.NewReader(str))
	decoder.DisallowUnknownFields()

	return emperror.Wrap(decoder.Decode(&pipeline.InstallSecretRequest{}), "invalid request")
}
