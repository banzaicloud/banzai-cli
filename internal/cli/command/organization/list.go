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
	"strings"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type listOptions struct {
	format string
}

// NewListCommand creates a new cobra.Command for `banzai organization list`.
func NewListCommand(banzaiCli cli.Cli) *cobra.Command {
	options := listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			options.format, _ = cmd.Flags().GetString("output")
			runList(banzaiCli, options)
		},
	}

	return cmd
}

func runList(banzaiCli cli.Cli, options listOptions) {
	orgs, _, err := banzaiCli.Client().OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		// TODO: review log usage
		log.Fatalf("could not list organizations: %v", err)
	}

	format.OrganizationWrite(banzaiCli.Out(), options.format, banzaiCli.Color(), orgs)

	// TODO: do this better
	if strings.ToLower(options.format) != "json" && strings.ToLower(options.format) != "yaml" {
		id := banzaiCli.Context().OrganizationID()
		for _, org := range orgs {
			if org.Id == id {
				// TODO: review log usage
				log.Infof("Organization %q (%d) is selected.", org.Name, org.Id)
			}
		}
	}
}
