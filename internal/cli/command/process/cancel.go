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
	"context"
	"fmt"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCancelCommand creates a new cobra.Command for `banzai process cancel`.
func NewCancelCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel processId",
		Short: "Cancel a process",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runCancel(banzaiCli, args)
		},
	}

	return cmd
}

func runCancel(banzaiCli cli.Cli, args []string) {
	id := args[0]
	_, err := banzaiCli.Client().ProcessesApi.CancelProcess(context.Background(), banzaiCli.Context().OrganizationID(), id)
	if err != nil {
		// TODO: review log usage
		log.Fatalf("could not cancel process: %v", err)
	}

	_, _ = fmt.Fprintf(banzaiCli.Out(), "process %q canceled\n", id)
}
