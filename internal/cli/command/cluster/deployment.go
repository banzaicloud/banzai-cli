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
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
	cmd.AddCommand(NewDeploymentDeleteCommand(banzaiCli))

	return cmd
}

// getClusterID returns the ID of the cluster selected by user either through command line flags
// or through the interactive prompt
func getClusterID(banzaiCli cli.Cli, orgID int32, options deploymentOptions) (int32, error) {
	var clusterID int32
	var err error

	if banzaiCli.Interactive() {
		clusterID, err = input.AskCluster(banzaiCli, orgID, options.clusterID, options.clusterName)
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
		return 0, errors.New("No cluster is specified. Use the --cluster or --cluster-name option or run the CLI in interactive mode")
	}

	if clusterID == 0 {
		return 0, errors.New("cluster could not be found")
	}

	return clusterID, nil
}

