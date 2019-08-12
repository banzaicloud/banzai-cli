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

package bucket

import (
	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

type getBucketOptions struct {
	name           string
	cloud          string
	location       string
	storageAccount string
}

// NewGetCommand creates a new cobra.Command for `banzai bucket get`.
func NewGetCommand(banzaiCli cli.Cli) *cobra.Command {
	o := getBucketOptions{}

	cmd := &cobra.Command{
		Use:     "get NAME [[--cloud=]CLOUD]]",
		Short:   "Get bucket",
		Args:    cobra.MaximumNArgs(2),
		Aliases: []string{"g"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if len(args) > 0 {
				o.name = args[0]
			}
			if len(args) > 1 {
				o.cloud = args[1]
			}

			if !banzaiCli.Interactive() {
				if o.name == "" {
					cmd.SilenceUsage = false
					return errors.New("NAME argument must be specified")
				}
				if o.cloud == "" {
					cmd.SilenceUsage = false
					return errors.New("CLOUD argument or --cloud flag must be specified")
				}
			} else if o.name != "" && o.cloud == "" {
				// Select cloud
				o.cloud, err = input.AskCloud()
				if err != nil {
					return err
				}
			}

			err = validateCloudAndLocation(banzaiCli, o.cloud, o.location)
			if err != nil {
				return err
			}

			if o.cloud == input.CloudProviderAzure && o.name != "" && o.storageAccount == "" {
				return errors.New("--storage-account must be specified for Azure")
			}

			return runGet(banzaiCli, o)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&o.cloud, "cloud", "", "", "Cloud provider for the bucket")
	flags.StringVarP(&o.location, "location", "l", "", "Location (e.g. us-central1) for the bucket")
	flags.StringVarP(&o.storageAccount, "storage-account", "", "", "Storage account for the bucket (must be specified for Azure)")

	return cmd
}

func runGet(banzaiCli cli.Cli, o getBucketOptions) error {
	var err error

	found, bucket, err := GetManagedBucket(banzaiCli, input.GetOrganization(banzaiCli), o.name, o.cloud, o.location, o.storageAccount)
	if err != nil {
		return err
	}

	if !found {
		if banzaiCli.OutputFormat() == output.OutputFormatDefault {
			log.Infof("No buckets were found")
		}
		return nil
	}

	format.DetailedBucketWrite(banzaiCli, bucket, bucket.Cloud)

	return nil
}
