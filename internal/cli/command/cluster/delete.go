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
	"fmt"

	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

type deleteOptions struct {
	force bool
}

func NewDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete NAME",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a cluster",
		Long:    "Delete a cluster. The cluster to delete is identified either by its name or the numerical ID. In case of interactive mode banzai CLI will prompt for a confirmation.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(banzaiCli, options, args)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.force, "force", "f", false, "Allow non-graceful cluster deletion")

	return cmd
}

func runDelete(banzaiCli cli.Cli, options deleteOptions, args []string) error {
	client := InitPipeline()
	orgId := GetOrgId(true)
	clusters, _, err := client.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		cli.LogAPIError("list clusters", err, orgId)
		return emperror.Wrap(err, "could not list clusters")
	}
	var id int32
	for _, cluster := range clusters {
		if cluster.Name == args[0] || fmt.Sprintf("%d", cluster.Id) == args[0] {
			id = cluster.Id
			break
		}
	}
	if id == 0 {
		return errors.New(fmt.Sprintf("cluster %q could not be found", args[0]))
	}

	if isInteractive() {
		if cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgId, id); err != nil {
			cli.LogAPIError("get cluster", err, id)
		} else {
			Out1(cluster, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt", "StatusMessage"})
		}
		confirmed := false
		survey.AskOne(&survey.Confirm{Message: "Do you want to DELETE the cluster?"}, &confirmed, nil)
		if !confirmed {
			return errors.New("deletion cancelled")
		}
	}
	if cluster, _, err := client.ClustersApi.DeleteCluster(context.Background(), orgId, id, &pipeline.DeleteClusterOpts{Force: optional.NewBool(options.force)}); err != nil {
		cli.LogAPIError("delete cluster", err, id)
		return emperror.Wrap(err, "failed to delete cluster")
	} else {
		log.Printf("Deleting cluster %v", cluster)
	}
	if cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgId, id); err != nil {
		cli.LogAPIError("get cluster", err, id)
	} else {
		Out1(cluster, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt", "StatusMessage"})
	}
	return nil
}
