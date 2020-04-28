// Copyright © 2020 Banzai Cloud
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

package process

import (
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/pkg/process"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type tailOptions struct {
	format string
}

// NewTailCommand creates a new cobra.Command for `banzai process tail`.
func NewTailCommand(banzaiCli cli.Cli) *cobra.Command {
	options := tailOptions{}

	cmd := &cobra.Command{
		Use:   "tail processId",
		Short: "Tail a process",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.format, _ = cmd.Flags().GetString("output")
			runTail(banzaiCli, options, args)
		},
	}

	return cmd
}

func runTail(banzaiCli cli.Cli, _ tailOptions, args []string) {
	err := process.TailProcess(banzaiCli, args[0])
	if err != nil {
		log.Fatal(err)
	}
}
