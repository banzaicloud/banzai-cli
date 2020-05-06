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

	"emperror.dev/errors"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/spf13/cobra"
)

type listOptions struct {
	format string
}

// NewListCommand creates a new cobra.Command for `banzai process list`.
func NewListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List processes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			options.format, _ = cmd.Flags().GetString("output")
			runList(banzaiCli, options)
		},
	}

	return cmd
}

func runList(banzaiCli cli.Cli, options listOptions) error {
	o := pipeline.ListProcessesOpts{
		Status: optional.NewInterface(pipeline.RUNNING),
	}

	processes, _, err := banzaiCli.Client().ProcessesApi.ListProcesses(context.Background(), banzaiCli.Context().OrganizationID(), &o)
	if err != nil {
		return errors.Wrap(err, "could not list processes")
	}

	format.ProcessWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), processes)

	return nil
}
