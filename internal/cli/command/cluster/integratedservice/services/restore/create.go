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
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createOptions struct {
	clustercontext.Context

	backupName string
}

func newRestoreCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Restore backup into a new cluster",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return runCreate(banzaiCli, options)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&options.backupName, "backupName", "", "", "Backup name")
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "create")

	return cmd
}

func runCreate(banzaiCli cli.Cli, options createOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	if options.backupName == "" {
		if banzaiCli.Interactive() {
			restore, err := askBackup(client, orgID, clusterID)
			if err != nil {
				return errors.WrapIf(err, "failed to ask restore")
			}

			options.backupName = restore.Name
		} else {
			return errors.NewWithDetails("invalid backup name", "backupName", options.backupName)
		}
	}

	_, _, err := client.ArkRestoresApi.CreateARKRestore(context.Background(), orgID, clusterID, pipeline.CreateRestoreRequest{
		BackupName: options.backupName,
	})
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to create restore", "clusterID", clusterID, "backupName", options.backupName)
	}

	log.Info("Restore started to create")

	return nil
}

func askBackup(client *pipeline.APIClient, orgID, clusterID int32) (*pipeline.BackupResponse, error) {
	backups, _, err := client.ArkBackupsApi.ListARKBackupsOfACluster(context.Background(), orgID, clusterID)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to list backups", "clusterID", clusterID)
	}

	backupOptions := make([]string, len(backups))
	for id, b := range backups {
		backupOptions[id] = b.Name
	}

	var selectedBackupName string
	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionSelect{
			QuestionInput: input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: "Restore cluster from this backup",
					Help:    "", // TODO (colin): need help msg??
				},
				Output: &selectedBackupName,
			},
			Options: backupOptions,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "failed to get bucket")
	}

	var selectedBucket pipeline.BackupResponse
	for idx, b := range backups {
		if b.Name == selectedBackupName || (selectedBackupName == "" && idx == 0) {
			selectedBucket = b
		}
	}

	return &selectedBucket, nil
}
