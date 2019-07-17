// Copyright © 2018 Banzai Cloud
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

package command

import (
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/bucket"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/cluster"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/controlplane"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/form"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/login"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/organization"
	"github.com/banzaicloud/banzai-cli/internal/cli/command/secret"
)

// AddCommands adds all the commands from cli/command to the root command
func AddCommands(cmd *cobra.Command, banzaiCli cli.Cli) {
	cmd.AddCommand(
		login.NewLoginCommand(banzaiCli),

		cluster.NewClusterCommand(banzaiCli),
		form.NewFormCommand(banzaiCli),
		organization.NewOrganizationCommand(banzaiCli),
		secret.NewSecretCommand(banzaiCli),
		controlplane.NewControlPlaneCommand(banzaiCli),
		bucket.NewBucketCommand(banzaiCli),
	)
}
