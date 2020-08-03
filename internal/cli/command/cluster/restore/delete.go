// Copyright © 2020 Banzai Cloud
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

package restore

import (
	"context"
	"strconv"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

type deleteOptions struct {
	clustercontext.Context
}

func newDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d", "remove"},
		Short:   "Delete logs of a restore job",
		Long:    "Delete logs of a restore job. Deleted jobs won't show up in the restore list.",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return runDelete(banzaiCli, options, args)
		},
	}
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "delete")

	return cmd
}

func runDelete(banzaiCli cli.Cli, options deleteOptions, args []string) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	var restoreID int32
	if len(args) > 0 {
		if id, err := strconv.ParseUint(args[0], 10, 64); err != nil {
			return errors.WrapIf(err, "failed to parse restoreID")
		} else {
			restoreID = int32(id)
		}
	}

	if restoreID == 0 {
		if banzaiCli.Interactive() {
			restore, err := askRestore(client, orgID, clusterID)
			if err != nil {
				return errors.WrapIf(err, "failed to ask restore")
			}

			restoreID = restore.Id
		} else {
			return errors.NewWithDetails("invalid restore ID", "restoreID", restoreID)
		}
	}

	_, _, err := client.ArkRestoresApi.DeleteARKRestore(context.Background(), orgID, clusterID, restoreID)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to delete restore", "clusterID", clusterID, "restoreID", restoreID)
	}

	log.Infof("Restore [%d] deleted successfully", restoreID)

	return nil
}
