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
	"fmt"
	"os"
	"strconv"
	"time"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/antihax/optional"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/format"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type createBucketsOptions struct {
	name     string
	cloud    string
	secretID string

	location string

	storageAccount string
	resourceGroup  string

	wait bool
}

// NewCreateCommand creates a new cobra.Command for `banzai bucket create`.
func NewCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	o := createBucketsOptions{}

	cmd := &cobra.Command{
		Use:     "create NAME [[--cloud=]CLOUD] [[--location=]LOCATION] [[--secret-id=]SECRET_ID]",
		Short:   "Create bucket",
		Long:    "Create object storage bucket on supported cloud providers",
		Args:    cobra.MaximumNArgs(4),
		Aliases: []string{"c", "create"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true

			if len(args) > 0 {
				o.name = args[0]
			}
			if len(args) > 1 {
				o.cloud = args[1]
			}
			if len(args) > 2 {
				o.location = args[2]
			}
			if len(args) > 3 {
				o.secretID = args[3]
			}

			if !banzaiCli.Interactive() {
				if o.name == "" {
					return errors.New("NAME must be specified")
				}
				if o.cloud == "" {
					return errors.New("CLOUD argument or --cloud flag must be specified")
				}
				if o.location == "" {
					return errors.New("LOCATION argument or --location flag must be specified")
				}
				if o.secretID == "" {
					return errors.New("SECRET_ID argument or --secret-id flag must be specified")
				}
			}

			if o.cloud == input.CloudProviderAzure {
				if o.storageAccount == "" {
					return errors.New("--storage-account must be specified for Azure")
				}
				if o.resourceGroup == "" {
					return errors.New("--resource-group must be specified for Azure")
				}
			}

			return runCreate(banzaiCli, o)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&o.cloud, "cloud", "c", "", "Cloud provider for the bucket")
	flags.StringVarP(&o.location, "location", "l", "", "Location for the bucket")
	flags.StringVarP(&o.secretID, "secret-id", "s", "", "Secret ID of the used secret to create the bucket")
	flags.StringVarP(&o.storageAccount, "storage-account", "", "", "Storage account for the bucket (must be specified for Azure)")
	flags.StringVarP(&o.resourceGroup, "resource-group", "", "", "Resource group for the bucket (must be specified for Azure)")
	flags.BoolVarP(&o.wait, "wait", "w", false, "Wait for bucket creation")

	return cmd
}

func runCreate(banzaiCli cli.Cli, o createBucketsOptions) error {
	var err error

	defaultBucketName := getDefaultBucketName()
	orgID := input.GetOrganization(banzaiCli)

	if o.cloud == "" {
		// Select cloud
		o.cloud, err = input.AskCloud()
		if err != nil {
			return err
		}
	} else {
		err = input.IsCloudProviderSupported(o.cloud)
		if err != nil {
			return err
		}
	}

	if o.name == "" {
		// Ask for bucket name
		err = survey.AskOne(&survey.Input{Message: "Bucket name:", Default: defaultBucketName}, &o.name, survey.WithValidator(input.BucketNameValidator(o.cloud)))
		if err != nil {
			return errors.WrapIf(err, "failed to select name")
		}
	} else {
		err = input.ValidateBucketName(o.cloud, o.name)
		if err != nil {
			return errors.WrapIf(err, "failed to validate bucket name")
		}
	}

	if o.location == "" {
		// Ask for location
		o.location, err = input.AskLocation(banzaiCli, o.cloud)
		if err != nil {
			return err
		}
	} else {
		err = input.IsLocationValid(banzaiCli, o.cloud, o.location)
		if err != nil {
			return err
		}
	}

	if o.secretID == "" {
		// Ask for secret
		o.secretID, err = input.AskSecret(banzaiCli, orgID, o.cloud)
		if err != nil {
			return err
		}
	}

	secret, _, err := banzaiCli.Client().SecretsApi.GetSecret(context.Background(), orgID, o.secretID)
	if err != nil {
		return errors.WrapIf(utils.ConvertError(errors.WithStack(err)), "could not get secret")
	}

	if secret.Type != o.cloud {
		return errors.Errorf("mismatching secret type '%s' for cloud '%s'", secret.Type, o.cloud)
	}

	if o.cloud == input.CloudProviderAzure {
		if banzaiCli.Interactive() {
			// Ask for storage account
			err = survey.AskOne(&survey.Input{Message: "Storage account:", Default: o.storageAccount}, &o.storageAccount)
			if err != nil {
				return errors.WrapIf(err, "failed to get storage account")
			}

			// Ask for resource group
			o.resourceGroup, err = input.AskResourceGroup(banzaiCli, orgID, o.secretID, o.resourceGroup)
			if err != nil {
				return errors.WrapIf(err, "failed to select resource group")
			}
		} else {
			err = input.IsResourceGroupValid(banzaiCli, orgID, o.secretID, o.resourceGroup)
			if err != nil {
				return err
			}
		}
	}

	request := getCreateBucketRequest(o)
	response, _, err := banzaiCli.Client().StorageApi.CreateObjectStoreBucket(context.Background(), orgID, request)
	if err != nil {
		return errors.WrapIf(utils.ConvertError(errors.WithStack(err)), "could not create bucket")
	}

	log.Infof("bucket create request accepted for %s on %s", response.Name, response.Cloud)

	done := !o.wait
	var getBucketOpts pipeline.GetBucketOpts
	if o.cloud == input.CloudProviderAzure {
		getBucketOpts.Location = optional.NewString(o.location)
		getBucketOpts.ResourceGroup = optional.NewString(o.resourceGroup)
		getBucketOpts.StorageAccount = optional.NewString(o.storageAccount)
	}

	for !done {
		log.Info("wait for response")
		time.Sleep(time.Duration(3) * time.Second)
		bucket, _, err := banzaiCli.Client().StorageApi.GetBucket(context.Background(), orgID, o.name, o.cloud, &getBucketOpts)
		if err != nil {
			return errors.WrapIf(utils.ConvertError(errors.WithStack(err)), "could not get bucket")
		}
		if bucket.Status != "CREATING" {
			format.DetailedBucketWrite(banzaiCli, ConvertBucketInfoToBucket(bucket), bucket.Cloud)
			done = true
		} else {
			time.Sleep(time.Duration(3) * time.Second)
		}
	}

	return nil
}

func getCreateBucketRequest(o createBucketsOptions) pipeline.CreateObjectStoreBucketRequest {
	// fill location for every provider since openapi doesn't generate those as pointers
	// TODO fix this
	properties := pipeline.CreateObjectStoreBucketProperties{
		Amazon:  &pipeline.CreateAmazonObjectStoreBucketProperties{Location: "n/a"},
		Google:  &pipeline.CreateGoogleObjectStoreBucketProperties{Location: "n/a"},
		Alibaba: &pipeline.CreateAlibabaObjectStoreBucketProperties{Location: "n/a"},
		Azure:   &pipeline.CreateAzureObjectStoreBucketProperties{Location: "n/a"},
		Oracle:  &pipeline.CreateOracleObjectStoreBucketProperties{Location: "n/a"},
	}

	switch o.cloud {
	case input.CloudProviderAlibaba:
		properties.Alibaba.Location = o.location
	case input.CloudProviderAmazon:
		properties.Amazon.Location = o.location
	case input.CloudProviderAzure:
		properties.Azure.Location = o.location
		properties.Azure.StorageAccount = o.storageAccount
		properties.Azure.ResourceGroup = o.resourceGroup
	case input.CloudProviderGoogle:
		properties.Google.Location = o.location
	case input.CloudProviderOracle:
		properties.Oracle.Location = o.location
	}

	return pipeline.CreateObjectStoreBucketRequest{
		SecretId:   o.secretID,
		Name:       o.name,
		Properties: properties,
	}
}

func getDefaultBucketName() string {
	prefix := os.Getenv("USER")
	if prefix == "" {
		prefix = "bucket"
	}
	return fmt.Sprintf("%s-%s", prefix, strconv.FormatInt(time.Now().UTC().Unix(), 10))
}
