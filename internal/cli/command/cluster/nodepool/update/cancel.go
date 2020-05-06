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
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCancelCommand creates a new cobra.Command for `banzai cluster nodepool update cancel`.
func NewCancelCommand(banzaiCli cli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel processId",
		Short: "Cancel a node pool update",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return runCancelUpdate(banzaiCli, args)
		},
	}

	return cmd
}

func runCancelUpdate(banzaiCli cli.Cli, args []string) error {
	processID := args[0]

	err := checkUpdateProcess(banzaiCli, processID)
	if err != nil {
		return err
	}

	_, err = banzaiCli.Client().ProcessesApi.CancelProcess(context.Background(), banzaiCli.Context().OrganizationID(), processID)
	if err != nil {
		return errors.WrapIf(err, "failed to cancel node pool update")
	}

	_, _ = fmt.Fprintf(banzaiCli.Out(), "update process %q canceled\n", processID)

	return nil
}
