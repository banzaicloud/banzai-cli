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
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createOptions struct {
	clustercontext.Context

	file     string
	wait     bool
	interval int
}

func NewCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a node pool for a given cluster",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return createNodePool(banzaiCli, options, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.file, "file", "f", "", "Node pool descriptor file")
	flags.BoolVarP(&options.wait, "wait", "w", false, "Wait for cluster creation")
	flags.IntVarP(&options.interval, "interval", "i", 10, "Interval in seconds for polling cluster status")

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "create")

	return cmd
}

//nolint:unparam
func createNodePool(banzaiCli cli.Cli, options createOptions, args []string) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()

	var out pipeline.NodePool

	err := options.Init()
	if err != nil {
		return err
	}

	clusterID := options.ClusterID()
	if clusterID == 0 {
		return errors.New("no clusters found")
	}

	filename, raw, err := utils.ReadFileOrStdin(options.file)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	log.Debugf("%d bytes read", len(raw))

	if err := validateNodePoolCreateRequest(raw); err != nil {
		return errors.WrapIf(err, "failed to parse create node pool request")
	}

	if err := utils.Unmarshal(raw, &out); err != nil {
		return errors.WrapIf(err, "failed to unmarshal create node pool request")
	}

	log.Debugf("create request: %#v", out)
	resp, err := client.ClustersApi.CreateNodePool(context.Background(), orgID, clusterID, out)
	if err != nil {
		cli.LogAPIError("create node pool", err, out)
		return errors.WrapIf(err, "failed to create nodepool")
	}
	if resp.StatusCode/100 != 2 {
		err := errors.NewWithDetails("Create nodepool failed with http status code", "status_code", resp.StatusCode)
		cli.LogAPIError("create node pool", err, out)
		return errors.WrapIf(err, "failed to create node pool")
	}

	return nil
}

func validateNodePoolCreateRequest(val interface{}) error {
	return nil
}
