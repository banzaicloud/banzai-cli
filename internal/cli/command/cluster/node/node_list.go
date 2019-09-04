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

package node

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

type nodeListOptions struct {
	clustercontext.Context
}

func NewNodeListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := nodeListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List cluster nodes",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNodeList(banzaiCli, options)
		},
	}
	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "node-list")

	return cmd
}

func runNodeList(banzaiCli cli.Cli, options nodeListOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()

	if err := options.Init(); err != nil {
		return err
	}

	id := options.ClusterID()

	nodes, _, err := client.ClustersApi.GetCluster(context.Background(), orgID, id)
	if err != nil {
		cli.LogAPIError("get cluster", err, id)
		log.Fatalf("could not get cluster: %v", err)
	}

	ctx := &output.Context{
		Out:    banzaiCli.Out(),
		Color:  banzaiCli.Color(),
		Format: "default",
		Fields: []string{"Name", "Status", "PoolName", "VCPU", "VCPUUsage", "Memory", "MemoryUsage", "InstanceType", "IsSpot"},
	}

	type Data struct {
		Name         string
		Status       string
		PoolName     string
		VCPU         string
		VCPUUsage    string
		Memory       string
		MemoryUsage  string
		InstanceType string
		IsSpot       bool
		Labels       []string
	}

	data := make([]Data, 0)

	for npName, np := range nodes.NodePools {
		for nodeName, rs := range np.ResourceSummary {
			data = append(data, Data{
				Name:         nodeName,
				Status:       rs.Status,
				PoolName:     npName,
				VCPU:         rs.Cpu.Capacity,
				VCPUUsage:    rs.Cpu.Request,
				Memory:       rs.Memory.Capacity,
				MemoryUsage:  rs.Memory.Request,
				InstanceType: np.InstanceType,
				IsSpot:       true,
			})
		}
	}

	err = output.Output(ctx, data)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
