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

	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/pipeline/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type listOptions struct {
	format     string
	secretType string
}

// NewListCommand creates a new cobra.Command for `banzai organization list`.
func NewListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runList(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.format, "format", "f", "default", "Output format (default|yaml|json)")
	flags.StringVarP(&options.secretType, "type", "t", "", "Filter list to the given type")

	return cmd
}

func runList(banzaiCli cli.Cli, options listOptions) {
	orgID := input.GetOrganization(banzaiCli)
	typeFilter := optional.EmptyString()
	if options.secretType != "" {
		typeFilter = optional.NewString(options.secretType)
	}
	secrets, _, err := banzaiCli.Client().SecretsApi.GetSecrets(context.Background(), orgID, &client.GetSecretsOpts{Type_: typeFilter})
	if err != nil {
		// TODO: review log usage
		log.Fatalf("could not list secrets: %v", err)
	}

	format.SecretWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), secrets)
}
