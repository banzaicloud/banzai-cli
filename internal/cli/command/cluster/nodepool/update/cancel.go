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
	"strings"

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
			return runCancelUpdate(banzaiCli, args)
		},
	}

	return cmd
}

func runCancelUpdate(banzaiCli cli.Cli, args []string) error {
	id := args[0]

	p, resp, err := banzaiCli.Client().ProcessesApi.GetProcess(context.Background(), banzaiCli.Context().OrganizationID(), id)
	if err != nil {
		cli.LogAPIError("cancel node pool update", err, resp)

		return errors.WrapIf(err, "failed to cancel node pool update")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.NewWithDetails("node pool update cancel failed with http status code", "status_code", resp.StatusCode)

		cli.LogAPIError("cancel node pool update", err, resp)

		return err
	}

	if !strings.HasSuffix(p.Type, "update-node-pool") {
		return errors.New("not a nodepool update process")
	}

	_, err = banzaiCli.Client().ProcessesApi.CancelProcess(context.Background(), banzaiCli.Context().OrganizationID(), id)
	if err != nil {
		// TODO: review log usage
		return errors.WrapIf(err, "failed to cancel node pool update")
	}

	_, _ = fmt.Fprintf(banzaiCli.Out(), "update process %q canceled\n", id)

	return nil
}
