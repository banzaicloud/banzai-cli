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
	"context"

	"github.com/antihax/optional"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/output"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type deleteBucketsOptions struct {
	name           string
	cloud          string
	location       string
	storageAccount string
}

// NewDeleteCommand creates a new cobra.Command for `banzai bucket delete`.
func NewDeleteCommand(banzaiCli cli.Cli) *cobra.Command {
	o := deleteBucketsOptions{}

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete bucket",
		Long:    "Delete Pipeline managed object storage bucket. Be aware, it also deletes the bucket at the cloud provider",
		Args:    cobra.MaximumNArgs(2),
		Aliases: []string{"d", "del"},
		RunE: func(cmd *cobra.Command, args []string) error {
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
			}

			err := validateCloudAndLocation(banzaiCli, o.cloud, o.location)
			if err != nil {
				return err
			}

			if o.cloud == input.CloudProviderAzure && o.storageAccount == "" {
				return errors.New("--storage-account must be specified for Azure")
			}

			return runDelete(banzaiCli, o)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&o.cloud, "cloud", "", "", "Cloud provider for the bucket")
	flags.StringVarP(&o.location, "location", "l", "", "Location (e.g. us-central1) for the bucket")
	flags.StringVarP(&o.storageAccount, "storage-account", "", "", "Storage account where the bucket resides (must be specified for Azure)")

	return cmd
}

func runDelete(banzaiCli cli.Cli, o deleteBucketsOptions) error {
	orgID := input.GetOrganization(banzaiCli)

	bucket, err := GetManagedBucket(banzaiCli, orgID, o.name, o.cloud, o.location, o.storageAccount)
	if err != nil {
		return err
	}

	if bucket.Cloud == "" && banzaiCli.OutputFormat() == output.OutputFormatDefault {
		if o.name != "" {
			log.Infof("no such bucket: %s", o.name)
		} else {
			log.Info("no buckets were found")
		}
		return nil
	}

	if banzaiCli.Interactive() {
		format.DetailedBucketWrite(banzaiCli, bucket, bucket.Cloud)

		confirmed := false
		survey.AskOne(&survey.Confirm{Message: "Do you want to DELETE the bucket?"}, &confirmed, nil)
		if !confirmed {
			return errors.New("deletion cancelled")
		}
	}

	var deleteOptions pipeline.DeleteObjectStoreBucketOpts
	deleteOptions.ResourceGroup = optional.NewString(bucket.ResourceGroup)
	deleteOptions.StorageAccount = optional.NewString(bucket.StorageAccount)
	deleteOptions.Location = optional.NewString(bucket.Location)

	_, err = banzaiCli.Client().StorageApi.DeleteObjectStoreBucket(context.Background(), orgID, bucket.Name, bucket.secretID, bucket.Cloud, &deleteOptions)
	if err != nil {
		return emperror.Wrap(utils.ConvertError(err), "could not delete bucket")
	}

	log.Infof("bucket '%s' successfully deleted", bucket.Name)

	return nil
}
