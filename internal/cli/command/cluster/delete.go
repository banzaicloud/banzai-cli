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

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	force bool
	clustercontext.Context
}

func NewDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete [--cluster=ID | [--cluster-name=]NAME]",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a cluster",
		Long:    "Delete a cluster. The cluster to delete is identified either by its name or the numerical ID. In case of interactive mode banzai CLI will prompt for a confirmation.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(banzaiCli, options, args)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.force, "force", "f", false, "Allow non-graceful cluster deletion")
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "delete")

	return cmd
}

func runDelete(banzaiCli cli.Cli, options deleteOptions, args []string) error {
	client := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()
	if err := options.Init(args...); err != nil {
		return err
	}
	id := options.ClusterID()

	if banzaiCli.Interactive() {
		if cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgId, id); err != nil {
			return errors.WrapIf(err, "failed to get cluster details")
		} else {
			format.ClusterWrite(banzaiCli, cluster)
		}
		confirmed := false
		survey.AskOne(&survey.Confirm{Message: "Do you want to DELETE the cluster?"}, &confirmed)
		if !confirmed {
			return errors.New("deletion cancelled")
		}
	}
	if cluster, err := client.ClustersApi.DeleteCluster(context.Background(), orgId, id, &pipeline.DeleteClusterOpts{Force: optional.NewBool(options.force)}); err != nil {
		cli.LogAPIError("delete cluster", err, id)
		return errors.WrapIf(err, "failed to delete cluster")
	} else {
		log.Printf("Deleting cluster %v", cluster)
	}
	if cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgId, id); err != nil {
		cli.LogAPIError("get cluster", err, id)
	} else {
		format.ClusterWrite(banzaiCli, cluster)
	}
	return nil
}
