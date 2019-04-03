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

package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/goph/emperror"
	"github.com/spf13/cobra"
)

type getOptions struct {
	format string
	name   string
	id     string
	hide   bool
}

// NewGetCommand creates a new cobra.Command for `banzai secret get`.
func NewGetCommand(banzaiCli cli.Cli) *cobra.Command {
	options := getOptions{}

	cmd := &cobra.Command{
		Use:     "get ([--name=]NAME | --id=ID)",
		Short:   "Get a secret",
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"g", "show", "sh"},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.format, _ = cmd.Flags().GetString("output")
			if len(args) == 1 {
				options.name = args[0]
			}
			return runGet(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.name, "name", "n", "", "Name of secret to get")
	flags.StringVarP(&options.id, "id", "i", "", "ID of secret to get")
	flags.BoolVarP(&options.hide, "hide", "H", false, "Hide secret contents in the output")

	return cmd
}

func runGet(banzaiCli cli.Cli, options getOptions) error {
	if options.id == "" && options.name == "" {
		return errors.New("specify either the name or the ID of the secret")
	}
	orgID := input.GetOrganization(banzaiCli)
	id := options.id
	if id == "" {
		secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &pipeline.GetSecretsOpts{})
		if err != nil {
			return emperror.Wrap(err, "could not list secrets")
		}
		for _, secret := range secrets {
			if secret.Name == options.name {
				id = secret.Id
				break
			}
		}
		if id == "" {
			return fmt.Errorf("can't find secret named %q", options.name)
		}
	}

	secret, _, err := banzaiCli.Client().SecretsApi.GetSecret(context.Background(), orgID, id)
	if err != nil {
		return emperror.Wrap(err, "could not get secret")
	}

	if options.hide {
		for i := range secret.Values {
			secret.Values[i] = "<hidden>"
		}
	}

	format.SecretWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), secret)
	return nil
}
