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

package backup

import (
	"context"
	"strconv"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
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
		Aliases: []string{"d", "remove", "rm"},
		Short:   "Delete the specified backup",
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

	enabled, err := isCommandEnabledForCluster(client, orgID, clusterID)
	if err != nil {
		return errors.WrapIf(err, "error during checking command availability")
	}

	if !enabled {
		return NotAvailableError{}
	}

	var backupID int32
	if len(args) > 0 {
		if id, err := strconv.ParseUint(args[0], 10, 64); err != nil {
			return errors.WrapIf(err, "failed to parse backupID")
		} else {
			backupID = int32(id)
		}
	}

	if backupID == 0 {
		if banzaiCli.Interactive() {
			backup, err := askBackupToDelete(client, orgID, clusterID)
			if err != nil {
				return errors.WrapIf(err, "failed to ask backup to delete")
			}

			backupID = backup.Id
		} else {
			return errors.NewWithDetails("invalid backup ID", "backupID", backupID)
		}
	}

	_, _, err = client.ArkBackupsApi.DeleteARKBackup(context.Background(), orgID, clusterID, backupID)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to delete backup", "clusterID", clusterID, "backupID", backupID)
	}

	log.Info("Backup deleted successfully")

	return nil
}

func askBackupToDelete(client *pipeline.APIClient, orgID, clusterID int32) (*pipeline.BackupResponse, error) {
	backups, _, err := client.ArkBackupsApi.ListARKBackupsOfACluster(context.Background(), orgID, clusterID)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to list backups", "clusterID", clusterID)
	}

	backupOptions := make([]string, len(backups))
	for id, b := range backups {
		backupOptions[id] = b.Name
	}

	var selectedBackupName string
	err = input.DoQuestions([]input.QuestionMaker{
		input.QuestionSelect{
			QuestionInput: input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: "Backup to delete",
				},
				Output: &selectedBackupName,
			},
			Options: backupOptions,
		},
	})
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get bucket to delete")
	}

	var selectedBackup pipeline.BackupResponse
	for idx, b := range backups {
		if b.Name == selectedBackupName || (selectedBackupName == "" && idx == 0) {
			selectedBackup = b
		}
	}

	return &selectedBackup, nil
}
