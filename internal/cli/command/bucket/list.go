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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
)

type listBucketsOptions struct {
	cloud    string
	location string
}

// NewListCommand creates a new cobra.Command for `banzai bucket list`.
func NewListCommand(banzaiCli cli.Cli) *cobra.Command {
	o := listBucketsOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List buckets",
		Args:    cobra.NoArgs,
		Aliases: []string{"l", "ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			err := validateCloudAndLocation(banzaiCli, o.cloud, o.location)
			if err != nil {
				return err
			}

			return runList(banzaiCli, o)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&o.cloud, "cloud", "", "", "Cloud provider where the bucket resides")
	flags.StringVarP(&o.location, "location", "l", "", "Location of the bucket")

	return cmd
}

func runList(banzaiCli cli.Cli, o listBucketsOptions) error {
	buckets, err := GetManagedBuckets(banzaiCli, input.GetOrganization(banzaiCli), o.cloud, o.location)
	if err != nil {
		return err
	}

	if len(buckets) < 1 {
		if banzaiCli.OutputFormat() == output.OutputFormatDefault {
			log.Info("No buckets were found")
		}
		return nil
	}

	for i, b := range buckets {
		if banzaiCli.OutputFormat() != output.OutputFormatDefault {
			continue
		}
		b.Name = b.formattedName()
		buckets[i] = b
	}

	format.BucketWrite(banzaiCli, buckets)

	return nil
}
