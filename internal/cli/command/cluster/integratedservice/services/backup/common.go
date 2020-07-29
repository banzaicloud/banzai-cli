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

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/spf13/cobra"
)

const (
	clusterRunningStatus = "RUNNING"
	clusterWarningStatus = "WARNING"
)

func NewBackupCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Enable the backup service",
		Long:  "Allows you to enable the backup service for the cluster, in order to create manual and scheduled automatic backups of the cluster, and also to restore from these backups. You must enable the backup service before it can be used. See the subcommands of `banzai cluster service backup` and `banzai cluster service restore` for details.",
	}

	cmd.AddCommand(
		newStatusCommand(banzaiCli),
		newEnableCommand(banzaiCli),
		newDisableCommand(banzaiCli),
		newListCommand(banzaiCli),
		newCreateCommand(banzaiCli),
		newDeleteCommand(banzaiCli),
	)

	return cmd
}

type NotAvailableError struct {
}

func (NotAvailableError) Error() string {
	return "This command is not available on this cluster. The following cloud providers are supported: Amazon, Azure, Google. The cluster must be in Running or Warning state to run this command."
}

func isCommandEnabledForCluster(client *pipeline.APIClient, orgID, clusterID int32) (bool, error) {
	response, _, err := client.ClustersApi.GetCluster(context.Background(), orgID, clusterID)
	if err != nil {
		return false, errors.WrapIfWithDetails(err, "failed to get cluster", "clusterID", clusterID)
	}

	switch response.Cloud {
	case amazonType, azureType, googleType:
		switch response.Status {
		case clusterRunningStatus, clusterWarningStatus:
			return true, nil
		default:
			return false, nil
		}
	default:
		return false, nil
	}
}
