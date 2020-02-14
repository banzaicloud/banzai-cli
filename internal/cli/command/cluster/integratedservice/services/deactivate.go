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

package services

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

type deactivateOptions struct {
	clustercontext.Context
}

type deactivateManager interface {
	ReadableName() string
	ServiceName() string
}

func newDeactivateCommand(banzaiCli cli.Cli, use string, mngr deactivateManager) *cobra.Command {
	options := deactivateOptions{}

	cmd := &cobra.Command{
		Use:     "deactivate",
		Aliases: []string{"disable", "off", "remove", "rm", "uninstall"},
		Short:   fmt.Sprintf("Deactivate the %s service of a cluster", mngr.ReadableName()),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runDeactivate(banzaiCli, mngr, options, args, use)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, fmt.Sprintf("deactivate %s cluster service of", mngr.ReadableName()))

	return cmd
}

func runDeactivate(
	banzaiCLI cli.Cli,
	m deactivateManager,
	options deactivateOptions,
	args []string,
	use string,
) error {
	if err := isServiceEnabled(context.Background(), banzaiCLI, use); err != nil {
		return errors.WrapIf(err, "failed to check service")
	}

	if err := options.Init(args...); err != nil {
		return errors.WrapIf(err, "failed to initialize options")
	}

	pipeline := banzaiCLI.Client()
	orgId := banzaiCLI.Context().OrganizationID()
	clusterId := options.ClusterID()

	resp, err := pipeline.IntegratedServicesApi.DeactivateIntegratedService(context.Background(), orgId, clusterId, m.ServiceName())
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("deactivate %s cluster service", m.ReadableName()), err, resp.Request)
		log.Fatalf("could not deactivate %s cluster service: %v", m.ReadableName(), err)
	}

	log.Infof("service %q started to deactivate", m.ReadableName())

	return nil
}
