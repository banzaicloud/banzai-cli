// Copyright © 2019 Banzai Cloud
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

package secret

import (
	"context"
	"fmt"

	"emperror.dev/errors"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	name bool
	ids  []string
}

// NewDeleteCommand creates a new cobra.Command for `banzai secret delete`.
func NewDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	options := deleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete [--name] secrets...",
		Short:   "Delete one or more secrets",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"d"},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.ids = args

			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			return runDelete(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.name, "name", "n", false, "Use name instead of secret ID")

	return cmd
}

func runDelete(banzaiCli cli.Cli, options deleteOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	ids := options.ids

	var failed bool

	// TODO: implement secret deletion in the API by secret name
	if options.name {
		secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{})
		if err != nil {
			return errors.WrapIf(err, "could not list secrets")
		}

		idMap := make(map[string]string)
		for _, id := range ids {
			idMap[id] = id
		}

		for _, secret := range secrets {
			if _, ok := idMap[secret.Name]; ok {
				idMap[secret.Name] = secret.Id
			}
		}

		ids = make([]string, 0, len(idMap))

		for name, id := range idMap {
			if id == name {
				failed = true

				_, _ = fmt.Fprintf(banzaiCli.Out(), "could not find secret named %q\n", name)

				continue
			}

			ids = append(ids, id)
		}
	}

	for _, id := range ids {
		_, err := banzaiCli.Client().SecretsApi.DeleteSecrets(context.Background(), orgID, id)
		if err != nil {
			failed = true

			_, _ = fmt.Fprintf(banzaiCli.Out(), "could not delete secret %q: %s\n", id, err.Error())

			continue
		}

		_, _ = fmt.Fprintf(banzaiCli.Out(), "secret %q deleted\n", id)
	}

	if failed {
		return errors.New("errors occurred during secret deletion")
	}

	return nil
}
