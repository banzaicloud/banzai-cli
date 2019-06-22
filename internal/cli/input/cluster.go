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

package input

import (
	"context"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"gopkg.in/AlecAivazis/survey.v1"
)

// AskCluster prompts the user for a cluster
// defaultClusterID is id of the cluster to be preselected in the show cluster list user can select a cluster from
// defaultClusterName is the name of the cluster to be preselected in the show cluster list user can select a cluster from
func AskCluster(banzaiCli cli.Cli, orgID int32, defaultClusterID int32, defaultClusterName string) (int32, error) {
	var clusterID int32

	clusters, _, err := banzaiCli.Client().ClustersApi.ListClusters(context.Background(), orgID)
	if err != nil {
		cli.LogAPIError("list clusters", err, orgID)
		return 0, emperror.Wrap(err, "could not list clusters")
	}

	if len(clusters) == 0 {
		return 0, errors.New("no clusters found in the current organization")
	}

	preSelectCluster := ""

	for _, cluster := range clusters {
		if cluster.Id == defaultClusterID || cluster.Name == defaultClusterName {
			preSelectCluster = cluster.Name
			break
		}
	}


	clusterSurveyInput := make([]string, len(clusters))
	for i, cluster := range clusters {
		clusterSurveyInput[i] = cluster.Name
	}

	clusterName := ""
	err = survey.AskOne(&survey.Select{Message: "Cluster", Options: clusterSurveyInput, Default: preSelectCluster}, &clusterName, survey.Required)
	if err != nil {
		return 0, emperror.Wrap(err, "error occurred while selecting cluster")
	}

	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			clusterID = cluster.Id
			break
		}
	}

	return clusterID, nil
}
