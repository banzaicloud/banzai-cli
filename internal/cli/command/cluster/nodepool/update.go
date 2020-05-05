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

package nodepool

import (
	"context"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/nodepool/update"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/banzaicloud/banzai-cli/pkg/process"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	clustercontext.Context

	file string
}

func NewUpdateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := updateOptions{}

	cmd := &cobra.Command{
		Use:     "update [NAME]",
		Aliases: []string{"u"},
		Short:   "Update a node pool (and related subcommands)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return updateNodePool(banzaiCli, options, args)
		},
	}

	flags := cmd.LocalFlags()

	flags.StringVarP(&options.file, "file", "f", "", "Node pool descriptor file")

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "update")

	cmd.AddCommand(
		update.NewCancelCommand(banzaiCli),
		update.NewTailCommand(banzaiCli),
	)

	return cmd
}

func updateNodePool(banzaiCli cli.Cli, options updateOptions, args []string) error {
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

	filename, raw, err := utils.ReadFileOrStdin(options.file)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	log.Debugf("%d bytes read", len(raw))

	var request pipeline.UpdateNodePoolRequest

	if err := utils.Unmarshal(raw, &request); err != nil {
		return errors.WrapIf(err, "failed to unmarshal update node pool request")
	}

	log.Debugf("update request: %#v", request)

	response, resp, err := client.ClustersApi.UpdateNodePool(context.Background(), orgID, clusterID, nodePoolName, request)
	if err != nil {
		cli.LogAPIError("update node pool", err, request)

		return errors.WrapIf(err, "failed to update node pool")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.NewWithDetails("node pool update failed with http status code", "status_code", resp.StatusCode)

		cli.LogAPIError("update node pool", err, request)

		return err
	}

	return process.TailProcess(banzaiCli, response.ProcessId)
}
