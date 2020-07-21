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

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

type deleteOptions struct {
	clustercontext.Context

	restoreID int32
}

func newDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"d", "remove"},
		Short:   "Delete logs of a restore job.",
		Long:    "Delete logs of a restore job. Deleted jobs won't show up in the restore list.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return runDelete(banzaiCli, options)
		},
	}
	flags := cmd.Flags()
	flags.Int32VarP(&options.restoreID, "restoreId", "", 0, "Restore ID")
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "delete")

	return cmd
}

func runDelete(banzaiCli cli.Cli, options deleteOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	if options.restoreID == 0 {
		if banzaiCli.Interactive() {
			restore, err := askRestore(client, orgID, clusterID)
			if err != nil {
				return errors.WrapIf(err, "failed to ask restore")
			}

			options.restoreID = restore.Id
		} else {
			return errors.NewWithDetails("invalid restore ID", "restoreID", options.restoreID)
		}
	}

	_, _, err := client.ArkRestoresApi.DeleteARKRestore(context.Background(), orgID, clusterID, options.restoreID)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to delete restore", "clusterID", clusterID, "restoreID", options.restoreID)
	}

	log.Infof("Restore [%d] deleted successfully", options.restoreID)

	return nil
}
