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

package dns

import (
	"context"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewDeactivateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deactivateOptions{}

	cmd := &cobra.Command{
		Use:     "deactivate",
		Aliases: []string{"disable", "off", "remove", "rm", "uninstall"},
		Short:   "Deactivate the DNS feature of a cluster",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runDeactivate(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "deactivate DNS cluster feature of")

	return cmd
}

type deactivateOptions struct {
	clustercontext.Context
}

func runDeactivate(banzaiCli cli.Cli, options deactivateOptions, args []string) error {
	if err := options.Init(args...); err != nil {
		return emperror.Wrap(err, "failed to initialize options")
	}

	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()

	resp, err := pipeline.ClusterFeaturesApi.DeactivateClusterFeature(context.Background(), orgId, clusterId, featureName)
	if err != nil {
		cli.LogAPIError("deactivate DNS cluster feature", err, resp.Request)
		log.Fatalf("could not deactivate DNS cluster feature: %v", err)
	}

	return nil
}
