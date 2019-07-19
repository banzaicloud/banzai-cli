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

package cluster

import (
	"context"

	pkgPipeline "github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type getOptions struct {
	clustercontext.Context
}

func NewGetCommand(banzaiCli cli.Cli) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:     "get [--cluster=ID | [--cluster-name=]NAME]",
		Aliases: []string{"g", "show"},
		Short:   "Get cluster details",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(banzaiCli, options, args)
		},
	}
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "get")

	return cmd
}

func runGet(banzaiCli cli.Cli, options getOptions, args []string) error {
	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()

	if err := options.Init(args...); err != nil {
		return err
	}

	id := options.ClusterID()
	cluster, _, err := pipeline.ClustersApi.GetCluster(context.Background(), orgId, id)
	if err != nil {
		cli.LogAPIError("get clusters", err, orgId)
		log.Fatalf("could not get clusters: %v", err)
	}

	type details struct {
		pkgPipeline.GetClusterStatusResponse
	}
	detailed := details{GetClusterStatusResponse: cluster}
	format.ClusterWrite(banzaiCli, detailed)
	return nil
}
