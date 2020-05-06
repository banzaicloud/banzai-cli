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

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCancelCommand creates a new cobra.Command for `banzai process cancel`.
func NewCancelCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel processId",
		Short: "Cancel a process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return runCancel(banzaiCli, args)
		},
	}

	return cmd
}

func runCancel(banzaiCli cli.Cli, args []string) error {
	id := args[0]
	_, err := banzaiCli.Client().ProcessesApi.CancelProcess(context.Background(), banzaiCli.Context().OrganizationID(), id)
	if err != nil {
		return errors.Wrap(err, "could not cancel process")
	}

	_, err = fmt.Fprintf(banzaiCli.Out(), "process %q canceled\n", id)
	return err
}
