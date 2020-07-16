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
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

type listOptions struct {
	clustercontext.Context
}

func newListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List restores", // TODO (colin): add desc
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return runList(banzaiCli, options)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "list")

	return cmd
}

func runList(banzaiCli cli.Cli, options listOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	restores, _, err := client.ArkRestoresApi.ListARKRestores(context.Background(), orgID, clusterID)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to list restores", "clusterID", clusterID)
	}

	type row struct {
		ID         int32
		Name       string
		BackupName string
		Status     string
	}

	table := make([]row, 0, len(restores))
	for _, r := range restores {
		table = append(table, row{
			ID:         r.Id,
			Name:       r.Name,
			BackupName: r.BackupName,
			Status:     r.Status,
		})
	}

	ctx := &output.Context{
		Out:    banzaiCli.Out(),
		Color:  banzaiCli.Color(),
		Format: banzaiCli.OutputFormat(),
		Fields: []string{"ID", "Name", "BackupName", "Status"},
	}

	if err := output.Output(ctx, table); err != nil {
		log.Fatal(err)
	}

	return nil
}
