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
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
)

func NewRestoreCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore the cluster from a backup",
	}

	cmd.AddCommand(
		newListCommand(banzaiCli),
		newResultCommand(banzaiCli),
		newDeleteCommand(banzaiCli),
		newRestoreCreateCommand(banzaiCli),
	)

	return cmd
}

func askRestore(client *pipeline.APIClient, orgID, clusterID int32) (*pipeline.RestoreResponse, error) {
	restores, _, err := client.ArkRestoresApi.ListARKRestores(context.Background(), orgID, clusterID)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to list restores", "clusterID", clusterID)
	}

	restoreOptions := make([]string, len(restores))
	for id, r := range restores {
		restoreOptions[id] = r.Name
	}

	var selectedRestoreName string
	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionSelect{
			QuestionInput: input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: "Restore job",
				},
				Output: &selectedRestoreName,
			},
			Options: restoreOptions,
		},
	}); err != nil {
		return nil, errors.WrapIf(err, "failed to ask restore")
	}

	var selectedRestore pipeline.RestoreResponse
	for idx, r := range restores {
		if r.Name == selectedRestoreName || (selectedRestoreName == "" && idx == 0) {
			selectedRestore = r
		}
	}

	return &selectedRestore, nil
}
