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
	var cmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

$ source <(banzai completion bash)

# To load completions for each session, execute once:
Linux:
  $ banzai completion bash > /etc/bash_completion.d/banzai
MacOS:
  $ banzai completion bash > /usr/local/etc/bash_completion.d/banzai

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ banzai completion zsh > "${fpath[1]}/_banzai"

# You will need to start a new shell for this setup to take effect.

Fish:

$ banzai completion fish | source

# To load completions for each session, execute once:
$ banzai completion fish > ~/.config/fish/completions/banzai.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
		},
	}
	return cmd
}
