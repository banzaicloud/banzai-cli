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

package integratedservice

import (
	"context"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type listOptions struct {
	clustercontext.Context
}

func NewListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List active (and pending) integrated services of a cluster",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "list services")

	return cmd
}

func runList(banzaiCli cli.Cli, options listOptions, args []string) error {
	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()

	if err := options.Init(args...); err != nil {
		return errors.WrapIf(err, "failed to initialize options")
	}

	clusterId := options.ClusterID()

	list, resp, err := pipeline.ClusterFeaturesApi.ListClusterFeatures(context.Background(), orgId, clusterId)
	if err != nil {
		cli.LogAPIError("list cluster services", err, resp.Request)
		log.Fatalf("could not list cluster services: %v", err)
	}

	type row struct {
		Name   string
		Status string
	}

	table := make([]row, 0, len(list))
	for name, details := range list {
		table = append(table, row{
			Name:   name,
			Status: details.Status,
		})
	}

	format.IntegratedServiceWrite(banzaiCli, table)

	return nil
}
