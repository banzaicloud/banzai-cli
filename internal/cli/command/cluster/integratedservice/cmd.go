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
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/dns"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/expiry"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/ingress"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/logging"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/monitoring"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/securityscan"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/integratedservice/services/vault"
)

func NewIntegratedServiceCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc", "integratedservice", "is"},
		Short:   "Manage cluster integrated services",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "list services")

	cmd.AddCommand(
		NewListCommand(banzaiCli),
		// NOTE: add integratedservice commands here
		services.NewServiceCommand(banzaiCli, "dns", dns.NewManager(banzaiCli)),
		services.NewServiceCommand(banzaiCli, "expiry", expiry.NewManager(banzaiCli)),
		services.NewServiceCommand(banzaiCli, "ingress", ingress.NewManager(banzaiCli)),
		services.NewServiceCommand(banzaiCli, "logging", logging.NewManager(banzaiCli)),
		services.NewServiceCommand(banzaiCli, "monitoring", monitoring.NewManager(banzaiCli)),
		services.NewServiceCommand(banzaiCli, "securityscan", securityscan.NewManager(banzaiCli)),
		services.NewServiceCommand(banzaiCli, "vault", vault.NewManager(banzaiCli)),
	)

	return cmd
}
