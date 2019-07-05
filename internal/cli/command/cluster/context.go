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
	"fmt"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

type Context interface {
	Init(...string) error
	ClusterID() int32
	ClusterName() string
}

type clusterContext struct {
	id        int32
	name      string
	banzaiCli cli.Cli
}

func NewClusterContext(cmd *cobra.Command, banzaiCli cli.Cli, verb string) Context {

	ctx := clusterContext{
		banzaiCli: banzaiCli,
	}
	flags := cmd.Flags()

	flags.Int32Var(&ctx.id, "cluster", 0, fmt.Sprintf("ID of cluster to %s", verb))
	viper.BindEnv(clusterIdKey, "BANZAI_CURRENT_CLUSTER_ID")
	flags.StringVar(&ctx.name, "cluster-name", "", fmt.Sprintf("Name of cluster to %s", verb))

	return &ctx
}

func (c *clusterContext) ClusterID() int32 {
	return c.id
}

func (c *clusterContext) ClusterName() string {
	return c.name
}

// Init completes the cluster context from the options, env vars, and if possible from the user
func (c *clusterContext) Init(args ...string) error {
	pipeline := c.banzaiCli.Client()
	orgId := c.banzaiCli.Context().OrganizationID()

	switch len(args) {
	case 0:
	case 1:
		c.name = args[0]
	default:
		return errors.New("invalid number of arguments")
	}

	if c.name == "" && c.id == 0 {
		c.id = viper.GetInt32(clusterIdKey)
	}

	if c.id != 0 {
		cluster, _, err := pipeline.ClustersApi.GetCluster(context.Background(), orgId, c.id)
		if err != nil {
			return emperror.Wrapf(err, "failed to retrieve cluster %d", c.id)
		}

		c.name = cluster.Name
		return nil
	}

	clusters, _, err := pipeline.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		return emperror.Wrap(err, "could not list clusters")
	}

	if len(clusters) == 0 {
		return emperror.Wrap(err, "there are no clusters in the organization")
	}

	if c.name == "" {
		if !c.banzaiCli.Interactive() {
			return errors.New("no cluster is selected; use the --cluster or --cluster-name option, or set the cluster.id config value")
		}

		clusterSlice := make([]string, len(clusters))
		for i, cluster := range clusters {
			clusterSlice[i] = cluster.Name
		}

		err := survey.AskOne(&survey.Select{Message: "Cluster:", Options: clusterSlice}, &c.name, survey.Required)
		if err != nil {
			return emperror.Wrap(err, "failed to select a cluster")
		}
	}

	for _, cluster := range clusters {
		if c.name == cluster.Name {
			c.id = cluster.Id
			return nil
		}
	}
	return errors.Errorf("could not find cluster named %q", c.name)
}
