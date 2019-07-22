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
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/spf13/cobra"
)

func NewDNSCommand(banzaiCli cli.Cli) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage cluster DNS feature",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runGet(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "manage DNS cluster feature of")

	cmd.AddCommand(
		NewDeactivateCommand(banzaiCli),
		NewGetCommand(banzaiCli),
	)

	return cmd
}
