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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

type disableOptions struct {
	clustercontext.Context
}

func newDisableCommand(banzaiCli cli.Cli) *cobra.Command {
	options := disableOptions{}

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable Backup service for the cluster. This disables the service, the backed up data won't be deleted. You have to do that manually.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return disableService(banzaiCli, options)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "disable")

	return cmd
}

func disableService(banzaiCli cli.Cli, options disableOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	_, _, err := client.ArkApi.DisableARK(context.Background(), orgID, clusterID)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to disable backup service", "clusterID", clusterID)
	}

	log.Info("Backup service is disabled")

	return nil
}
