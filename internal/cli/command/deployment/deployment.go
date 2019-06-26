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
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/ghodss/yaml"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type deploymentOptions struct {
	clusterName string
	clusterID   int32

	// https://github.com/golang/lint/issues/433
	// nolint: structcheck
	format string
}

// NewDeploymentCommand returns a `*cobra.Command` for `banzai cluster deployment` subcommands.
func NewDeploymentCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "deployment",
		Aliases:       []string{"deployments", "deploy"},
		Short:         "Manage deployments",
	}

	cmd.AddCommand(NewDeploymentListCommand(banzaiCli))
	cmd.AddCommand(NewDeploymentGetCommand(banzaiCli))
	cmd.AddCommand(NewDeploymentCreateCommand(banzaiCli))
	cmd.AddCommand(NewDeploymentUpdateCommand(banzaiCli))
	cmd.AddCommand(NewDeploymentDeleteCommand(banzaiCli))

	return cmd
}

// getClusterID returns the ID of the cluster selected by user either through command line flags
// or through the interactive prompt
func getClusterID(banzaiCli cli.Cli, orgID int32, options deploymentOptions) (int32, error) {
	var clusterID int32
	var err error

	if options.clusterID == 0 && banzaiCli.Context().ClusterID() != 0 {
		options.clusterID = banzaiCli.Context().ClusterID()
	}

	if banzaiCli.Interactive() &&  options.clusterID == 0 && options.clusterName == "" {
		clusterID, err = input.AskCluster(banzaiCli, orgID)
		if err != nil {
			return 0, emperror.Wrap(err, "could not ask for a cluster")
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
			return 0, emperror.Wrap(err, "could not list clusters")
		}

		for _, cluster := range clusters {
			if cluster.Name == options.clusterName {
				clusterID = cluster.Id
				break
			}
		}
	} else {
		return 0, errors.New("No cluster is specified. Use the --cluster or --cluster-name option or select cluster in interactive mode")
	}

	if clusterID == 0 {
		return 0, errors.New("cluster could not be found")
	}

	return clusterID, nil
}


// getReleaseName returns the release name passed in if exists.
// If the release name passed in is empty than prompts the user interactively
// for a release name.
func getReleaseName(banzaiCli cli.Cli, orgID, clusterID int32, releaseName string) (string, error) {
	var name string
	var err error

	if banzaiCli.Interactive() && releaseName == "" {
		name, err = input.AskDeployment(banzaiCli, orgID, clusterID)
		if err != nil {
			return "", emperror.Wrap(err, "could not ask for a deployment")
		}
	} else if releaseName != "" {
		_, err = banzaiCli.Client().DeploymentsApi.HelmDeploymentStatus(context.Background(), orgID, clusterID, releaseName)
		if err != nil {
			return "", emperror.Wrapf(err, "deployment with release name %q could not be found", releaseName)
		}

		name = releaseName
	} else {
		return "", errors.New("No release name is specified!")
	}


	return name, nil
}

func buildCreateUpdateDeploymentRequest(banzaiCli cli.Cli, fileName string) (*pipeline.CreateUpdateDeploymentRequest, error) {
	var raw []byte
	var err error

	if banzaiCli.Interactive() {
		for {
			if fileName == "" {
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
					log.Errorf("failed to read file %q: %v", fileName, err)
					fileName = "" // reset fileName so that we can ask for one

					continue
				}
			}
			break
		}
	}

	if fileName != "" && fileName != "-" && raw == nil {
		if raw, err = ioutil.ReadFile(fileName); err != nil {
			return nil, emperror.Wrapf(err, "failed to read from file %q", fileName)
		}
	} else {
		if raw, err = ioutil.ReadAll(os.Stdin); err != nil {
			return nil, emperror.Wrap(err, "failed to read from stdin",)
		}
	}

	req, err := unmarshalCreateUpdateDeploymentRequest(raw)
	if err != nil {
		return nil, emperror.Wrap(err, "could not parse deployment request")
	}

	return req, nil


}

func unmarshalCreateUpdateDeploymentRequest(data []byte) (*pipeline.CreateUpdateDeploymentRequest, error) {
	req := pipeline.CreateUpdateDeploymentRequest{}

	// try json
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var errJSON error
	if errJSON = decoder.Decode(&req); errJSON == nil {
		return &req, nil
	}

	// try yaml
	var errYaml error
	if errYaml = yaml.Unmarshal(data, &req); errYaml == nil {
		return &req, nil
	}

	return nil, errors.Errorf("JSON unmarshal failed: %v, YAML unmarshal failed: %v ", errJSON, errYaml)
}

