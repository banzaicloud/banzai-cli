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

package organization

import (
	"context"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type selectOptions struct {
	organization string
}

// NewSelectCommand creates a new cobra.Command for `banzai organization select`.
func NewSelectCommand(banzaiCli cli.Cli) *cobra.Command {
	options := selectOptions{}

	cmd := &cobra.Command{
		Use:   "select [ORG NAME]",
		Short: "Select and organization",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				options.organization = args[0]
			}

			runSelect(banzaiCli, options)
		},
	}

	return cmd
}

func runSelect(banzaiCli cli.Cli, options selectOptions) {
	if options.organization != "" {
		orgs, _, err := banzaiCli.Client().OrganizationsApi.ListOrgs(context.Background())
		if err != nil {
			// TODO: review log usage
			cli.LogAPIError("list organizations", err, nil)
			log.Fatal("")
		}

		for _, org := range orgs {
			if org.Name == options.organization {
				banzaiCli.Context().SetOrganizationID(org.Id)

				return
			}
		}

		// TODO: review log usage
		log.Errorf("Could not find organization %q", options.organization)

		return
	}

	organizationID := input.AskOrganization(banzaiCli)

	banzaiCli.Context().SetOrganizationID(organizationID)
}
