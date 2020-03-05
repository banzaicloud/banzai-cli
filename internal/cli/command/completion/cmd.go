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

package completion

import (
	"os"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCompletionCommand returns a cobra command for `completion` subcommands.
func NewCompletionCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion SHELL",
		Short: "Generates shell completion scripts",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "bash",
			Short: "Generates bash completion scripts",
			Long: `To load completion run
			
			. <(banzai completion bash)
			`,
			Run: func(c *cobra.Command, args []string) {
				c.GenBashCompletion(os.Stdout)
			},
		},
		&cobra.Command{
			Use:   "zsh",
			Short: "Generates zsh completion scripts",
			Long: `To load completion run
				
				. <(banzai completion zsh)
				`,
			Run: func(c *cobra.Command, args []string) {
				c.GenZshCompletion(os.Stdout)
			},
		},
	)
	return cmd
}
