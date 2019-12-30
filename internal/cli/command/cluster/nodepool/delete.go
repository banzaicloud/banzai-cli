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

package nodepool

import (
	"context"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	clustercontext.Context

	nodePoolName string
}

func NewDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	o := deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete [NODE_POOL_NAME]",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a node pool for a given cluster",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteNodePool(banzaiCli, o, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&o.nodePoolName, "node-pool-name", o.nodePoolName, "Node pool name")

	o.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "delete")

	return cmd
}

func deleteNodePool(banzaiCli cli.Cli, options deleteOptions, args []string) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()

	err := options.Init()
	if err != nil {
		return err
	}

	clusterID := options.ClusterID()
	if clusterID == 0 {
		return errors.New("no clusters found")
	}

	var nodePoolName string
	if len(args) > 0 {
		nodePoolName = args[0]
	}
	if nodePoolName == "" && options.nodePoolName != "" {
		nodePoolName = options.nodePoolName
	}

	if nodePoolName == "" && !banzaiCli.Interactive() {
		return errors.New("no node pool is selected; use the --node-pool-name option or add node pool name as an argument")
	}

	log.Debugf("delete request: %s", nodePoolName)
	resp, err := client.ClustersApi.DeleteNodePool(context.Background(), orgID, clusterID, nodePoolName)
	if err != nil {
		cli.LogAPIError("delete node pool", err, nodePoolName)
		return errors.WrapIf(err, "failed to delete node pool")
	}
	if resp.StatusCode/100 != 2 {
		err := errors.NewWithDetails("Delete node pool failed with http status code", "status_code", resp.StatusCode, "nodepool", nodePoolName)
		cli.LogAPIError("delete node pool", err, nodePoolName)
		return errors.WrapIf(err, "failed to delete node pool")
	}

	return nil
}
