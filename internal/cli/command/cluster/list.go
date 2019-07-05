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

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type listOptions struct {
}

func NewListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List clusters",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(banzaiCli, options)
		},
	}

	return cmd
}

func runList(banzaiCli cli.Cli, options listOptions) error {
	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()

	clusters, _, err := pipeline.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		cli.LogAPIError("list clusters", err, orgId)
		log.Fatalf("could not list clusters: %v", err)
	}

	format.ClustersWrite(banzaiCli, clusters)
	return nil
}
