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
}

func NewDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete [NAME]",
		Aliases: []string{"del", "rm"},
		Short:   "Delete a node pool",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return deleteNodePool(banzaiCli, options, args)
		},
		Hidden: true,
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "delete")

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

	nodePoolName := args[0]

	log.Debugf("delete request: %s", nodePoolName)
	resp, err := client.ClustersApi.DeleteNodePool(context.Background(), orgID, clusterID, nodePoolName)
	if err != nil {
		cli.LogAPIError("delete node pool", err, nodePoolName)

		return errors.WrapIf(err, "failed to delete node pool")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.NewWithDetails("node pool deletion failed with http status code", "status_code", resp.StatusCode, "nodePool", nodePoolName)

		cli.LogAPIError("delete node pool", err, nodePoolName)

		return err
	}

	return nil
}
