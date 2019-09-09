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

package features

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type deactivateOptions struct {
	clustercontext.Context
}

type DeactivateManager interface {
	GetName() string
}

func DeactivateCommandFactory(banzaiCli cli.Cli, manager DeactivateManager, name string) *cobra.Command {
	options := deactivateOptions{}

	cmd := &cobra.Command{
		Use:     "deactivate",
		Aliases: []string{"disable", "off", "remove", "rm", "uninstall"},
		Short:   fmt.Sprintf("Deactivate the %s feature of a cluster", name),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runDeactivate(banzaiCli, manager, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, fmt.Sprintf("deactivate %s cluster feature of", name))

	return cmd
}

func runDeactivate(
	banzaiCLI cli.Cli,
	m DeactivateManager,
	options deactivateOptions,
	args []string,
) error {
	if err := options.Init(args...); err != nil {
		return errors.WrapIf(err, "failed to initialize options")
	}

	pipeline := banzaiCLI.Client()
	orgId := banzaiCLI.Context().OrganizationID()
	clusterId := options.ClusterID()

	resp, err := pipeline.ClusterFeaturesApi.DeactivateClusterFeature(context.Background(), orgId, clusterId, m.GetName())
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("deactivate %s cluster feature", m.GetName()), err, resp.Request)
		log.Fatalf("could not deactivate %s cluster feature: %v", m.GetName(), err)
	}

	log.Infof("feature %q started to deactivate", m.GetName())

	return nil
}
