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

package update

import (
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/pkg/process"
	"github.com/spf13/cobra"
)

// NewTailCommand creates a new cobra.Command for `banzai cluster nodepool update tail`.
func NewTailCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tail processId",
		Short: "Tail a node pool update",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTail(banzaiCli, args)
		},
	}

	return cmd
}

func runTail(banzaiCli cli.Cli, args []string) error {
	processID := args[0]

	err := checkUpdateProcess(banzaiCli, processID)
	if err != nil {
		return err
	}

	return process.TailProcess(banzaiCli, processID)
}
