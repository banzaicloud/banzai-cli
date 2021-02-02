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

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

// Bucket describes an object store bucket managed by Pipeline
type Bucket struct {
	// the name of the object storage / bucket
	Name string `json:"name"`
	// true if the bucket has been created via pipeline
	Managed bool `json:"managed"`
	// cloud provider where the bucket resides
	Cloud string `json:"cloud"`
	// location where the bucket resides
	Location string `json:"location"`
	// notes for the bucket
	Notes string `json:"notes,omitempty" yaml:"notes,omitempty"`
	// the status of the bucket
	Status string `json:"status"`
	// the reason for the error status
	StatusMessage string `json:"statusMessage,omitempty" yaml:"statusMessage,omitempty"`

	// Azure property
	StorageAccount string `json:"storageAccount,omitempty" yaml:"storageAccount,omitempty"`
	// Azure property
	ResourceGroup string `json:"resourceGroup,omitempty" yaml:"resourceGroup,omitempty"`

	secretID   string
	secretName string
}

// GetManagedBuckets gets managed buckets from Pipeline
func GetManagedBuckets(banzaiCli cli.Cli, orgID int32, cloud, location string) ([]Bucket, error) {
	managedBuckets, _, err := banzaiCli.Client().StorageApi.ListObjectStoreBuckets(context.Background(), orgID, &pipeline.ListObjectStoreBucketsOpts{})
	if err != nil {
		err = utils.ConvertError(errors.WithStack(err))
		return nil, errors.WrapIf(err, "could not get buckets")
	}

	buckets := make([]Bucket, 0)
	for _, bucket := range ConvertBucketInfoToBuckets(managedBuckets) {
		if cloud != "" && cloud != bucket.Cloud {
			continue
		}
		if location != "" && location != bucket.Location {
			continue
		}
		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

// GetManagedBucket gets a managed bucket from Pipeline
func GetManagedBucket(banzaiCli cli.Cli, orgID int32, name, cloud, location, storageAccount string) (bool, Bucket, error) {
	var selectedBucket Bucket
	var err error

	buckets, err := GetManagedBuckets(banzaiCli, orgID, cloud, location)
	if err != nil {
		return false, selectedBucket, err
	}

	if len(buckets) < 1 && name == "" {
		return false, selectedBucket, nil
	}

	if name == "" && banzaiCli.Interactive() {
		bucketSlice := make([]string, 0)
		bucketNames := make(map[string]Bucket, 0)
		for _, bucket := range buckets {
			bucketSlice = append(bucketSlice, bucket.GetNameForSelection())
			bucketNames[bucket.GetNameForSelection()] = bucket
		}

		var selectedName string
		err = survey.AskOne(&survey.Select{Message: getTitlesForBucketSelection(), Options: bucketSlice}, &selectedName, survey.WithValidator(survey.Required))
		if err != nil {
			return false, selectedBucket, errors.WrapIf(err, "failed to select bucket")
		}
		selectedBucket = bucketNames[selectedName]
	} else {
		for _, bucket := range buckets {
			if bucket.Name == name && bucket.Cloud == cloud &&
				(cloud != input.CloudProviderAzure || storageAccount == bucket.StorageAccount) {
				selectedBucket = bucket
				break
			}
		}
	}

	if selectedBucket.Cloud == "" {
		return false, selectedBucket, errors.Errorf("no such bucket: %s", name)
	}

	return true, selectedBucket, nil
}

// ConvertBucketInfoToBucket converts pipeline.BucketInfo to Bucket
func ConvertBucketInfoToBucket(bucket pipeline.BucketInfo) Bucket {
	return Bucket{
		Name:          bucket.Name,
		Managed:       bucket.Managed,
		Cloud:         bucket.Cloud,
		Location:      bucket.Location,
		Notes:         bucket.Notes,
		Status:        bucket.Status,
		StatusMessage: bucket.StatusMessage,

		StorageAccount: bucket.Aks.StorageAccount,
		ResourceGroup:  bucket.Aks.ResourceGroup,

		secretID:   bucket.Secret.Id,
		secretName: bucket.Secret.Name,
	}
}

// ConvertBucketInfoToBuckets converts an array of []pipeline.BucketInfo to []Bucket
func ConvertBucketInfoToBuckets(bucketInfos []pipeline.BucketInfo) []Bucket {
	buckets := make([]Bucket, len(bucketInfos))
	for i, b := range bucketInfos {
		buckets[i] = ConvertBucketInfoToBucket(b)
	}

	return buckets
}

// GetNameForSelection gets a specially formatted name for interactive bucket selection
func (bucket Bucket) GetNameForSelection() string {
	if bucket.StorageAccount == "" {
		bucket.StorageAccount = "-"
	}
	return fmt.Sprintf("%-30s %-15s %-15s %-15s %-15s", bucket.Name, bucket.Cloud, bucket.Location, bucket.StorageAccount, bucket.Status)
}

func getTitlesForBucketSelection() string {
	return fmt.Sprintf("%-30s %-15s %-15s %-15s\n", "Bucket", "Cloud", "Location", "StorageAccount")
}

func (bucket Bucket) formattedName() string {
	if bucket.StorageAccount != "" {
		bucket.Name = bucket.Name + " (" + bucket.StorageAccount + ")"
	}

	return bucket.Name
}

func validateCloudAndLocation(banzaiCli cli.Cli, cloud, location string) error {
	if cloud != "" {
		err := input.IsCloudProviderSupported(cloud)
		if err != nil {
			return err
		}

		if location != "" {
			err = input.IsLocationValid(banzaiCli, cloud, location)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
