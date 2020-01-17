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

package form

import (
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewFormCommand returns a cobra command for `form` subcommands.
func NewFormCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "form",
		Short:      "Open forms from config, persist provided values and generate templates",
		Deprecated: "This command and subcommands will be removed later.",
	}

	cmd.AddCommand(NewOpenCommand(banzaiCli))
	cmd.AddCommand(NewTemplateCommand(banzaiCli))
	cmd.AddCommand(NewMigrateCommand(banzaiCli))

	return cmd
}
