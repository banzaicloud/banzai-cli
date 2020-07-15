// Copyright © 2020 Banzai Cloud
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

package backup

import (
	"context"
	"encoding/json"

	"emperror.dev/errors"
	"github.com/antihax/optional"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
)

const (
	amazonType = "amazon"
	azureType  = "azure"
	googleType = "google"
)

const (
	scheduleDailyLabel   = "daily"
	scheduleDailyValue   = "0 12 * * *"
	scheduleWeeklyLabel  = "weekly"
	scheduleWeeklyValue  = "0 12 * * 0"
	scheduleMonthlyLabel = "monthly"
	scheduleMonthlyValue = "0 12 1 * *"
)

const (
	ttl1DayLabel  = "1 day"
	ttl1DayValue  = "24h"
	ttl2DaysLabel = "2 days"
	ttl2DaysValue = "48h"
	ttl1WeekLabel = "1 week"
	ttl1WeekValue = "168h"
)

const (
	providerAmazonS3Label  = "Amazon S3"
	providerGoogleGCSLabel = "Google Cloud Storage"
	providerAzureLabel     = "Azure Blob Storage"
)

type bucketInfo struct {
	provider string
	secretID string
	name     string
}

type enableOptions struct {
	clustercontext.Context

	filePath string
}

func newEnableCommand(banzaiCli cli.Cli) *cobra.Command {
	options := enableOptions{}

	cmd := &cobra.Command{
		Use:     "enable",
		Aliases: []string{"e", "activate", "on"},
		Short:   "Enable Backup service", // TODO (colin): add desc
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return enableService(banzaiCli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Enable backup service specification file")

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "enable")

	return cmd
}

func enableService(banzaiCli cli.Cli, options enableOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	response, _, err := client.ArkApi.CheckARKStatusGET(context.Background(), orgID, clusterID)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to check backup status", "clusterID", clusterID)
	}

	if response.Enabled {
		return errors.New("backup service already enabled")
	}

	var request pipeline.EnableArkRequest
	if options.filePath == "" && banzaiCli.Interactive() {
		if request, err = buildEnableRequestInteractively(banzaiCli); err != nil {
			return errors.WrapIf(err, "failed to build enable backup service request interactively")
		}
	} else {
		if err = readEnableReqFromFileOrStdin(options.filePath, &request); err != nil {
			return errors.WrapIf(err, "failed to read enable backup service specification")
		}
	}

	log.Infof("Backup service is started to enable for [%d] cluster", clusterID)

	_, _, err = client.ArkApi.EnableARK(context.Background(), orgID, clusterID, request)
	if err != nil {
		return errors.WrapIf(err, "failed to enable backup service")
	}

	log.Infof("Backup service is enabled for [%d] cluster", clusterID)

	return nil
}

func readEnableReqFromFileOrStdin(filePath string, req *pipeline.EnableArkRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func buildEnableRequestInteractively(banzaiCli cli.Cli) (pipeline.EnableArkRequest, error) {
	var scheduleLabel string
	var ttlLabel string

	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionSelect{
			QuestionInput: input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: "Schedule", // TODO (colin): add message
					Help:    "",         // TODO (colin): need help message?
				},
				DefaultValue: scheduleDailyLabel,
				Output:       &scheduleLabel,
			},
			Options: []string{scheduleDailyLabel, scheduleWeeklyLabel, scheduleMonthlyLabel},
		},
		input.QuestionSelect{
			QuestionInput: input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: "Keep backups for", // TODO (colin): add message
					Help:    "",                 // TODO (colin): need help message?
				},
				DefaultValue: ttl1DayLabel,
				Output:       &ttlLabel,
			},
			Options: []string{ttl1DayLabel, ttl2DaysLabel, ttl1WeekLabel},
		},
	}); err != nil {
		return pipeline.EnableArkRequest{}, errors.WrapIf(err, "error during getting enable options")
	}

	var selectedSchedule string
	switch scheduleLabel {
	case scheduleDailyLabel:
		selectedSchedule = scheduleDailyValue
	case scheduleWeeklyLabel:
		selectedSchedule = scheduleWeeklyValue
	case scheduleMonthlyLabel:
		selectedSchedule = scheduleMonthlyValue
	default:
		return pipeline.EnableArkRequest{}, errors.New("not supported schedule")
	}

	var selectedTTL string
	switch ttlLabel {
	case ttl1DayLabel:
		selectedTTL = ttl1DayValue
	case ttl2DaysLabel:
		selectedTTL = ttl2DaysValue
	case ttl1WeekLabel:
		selectedTTL = ttl1WeekValue
	}

	bucketInfo, err := askBucketOptions(banzaiCli)
	if err != nil {
		return pipeline.EnableArkRequest{}, errors.WrapIf(err, "failed to get bucket")
	}

	return pipeline.EnableArkRequest{
		Cloud:      bucketInfo.provider,
		BucketName: bucketInfo.name,
		Schedule:   selectedSchedule,
		Ttl:        selectedTTL,
		SecretId:   bucketInfo.secretID,
	}, nil
}

func askBucketOptions(banzaiCli cli.Cli) (*bucketInfo, error) {
	selectedProvider, err := askBucketProvider()
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get bucket provider")
	}

	var secretAndBucketType string
	switch selectedProvider {
	case providerAmazonS3Label:
		secretAndBucketType = amazonType
	case providerAzureLabel:
		secretAndBucketType = azureType
	case providerGoogleGCSLabel:
		secretAndBucketType = googleType
	default:
		return nil, errors.NewWithDetails("not supported bucket provider", "provider", selectedProvider)
	}

	secretID, err := askSecret(banzaiCli, secretAndBucketType)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get secret")
	}

	bucket, err := askBucket(banzaiCli, secretAndBucketType, secretID)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get bucket")
	}

	return &bucketInfo{
		provider: secretAndBucketType,
		secretID: secretID,
		name:     bucket.Name,
	}, nil
}

func askBucket(banzaiCLI cli.Cli, bucketType, secretID string) (*pipeline.BucketInfo, error) {
	orgID := banzaiCLI.Context().OrganizationID()
	buckets, _, err := banzaiCLI.Client().StorageApi.ListObjectStoreBuckets(
		context.Background(),
		orgID,
		&pipeline.ListObjectStoreBucketsOpts{
			SecretId:  optional.NewString(secretID),
			CloudType: optional.NewString(bucketType),
		},
	)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "failed to get bucket(s)", "secretID", secretID, "cloud", bucketType)
	}

	if len(buckets) == 0 {
		return nil, errors.New("there's no buckets")
	}

	var defaultBucketName string
	secretOptions := make([]string, len(buckets))
	for i, b := range buckets {
		secretOptions[i] = b.Name
		if i == 0 {
			defaultBucketName = b.Name
		}
	}

	var bucketName string
	if err := input.DoQuestions([]input.QuestionMaker{input.QuestionSelect{
		QuestionInput: input.QuestionInput{
			QuestionBase: input.QuestionBase{
				Message: "Bucket name:",
			},
			DefaultValue: defaultBucketName,
			Output:       &bucketName,
		},
		Options: secretOptions,
	}}); err != nil {
		return nil, errors.WrapIf(err, "error during getting secret")
	}

	var bucket pipeline.BucketInfo
	for _, b := range buckets {
		if b.Name == bucketName {
			bucket = b
			return &bucket, nil
		}
	}

	return nil, errors.New("missing bucket")
}

func askBucketProvider() (string, error) {
	var selectedProviderName string
	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionSelect{
			QuestionInput: input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: "Select storage provider:", // TODO (colin): add message
				},
				DefaultValue: providerAmazonS3Label,
				Output:       &selectedProviderName,
			},
			Options: []string{providerAmazonS3Label, providerAzureLabel, providerGoogleGCSLabel},
		},
	}); err != nil {
		return "", errors.WrapIf(err, "error during getting provider")
	}

	return selectedProviderName, nil
}

func askSecret(banzaiCLI cli.Cli, secretType string) (string, error) {
	orgID := banzaiCLI.Context().OrganizationID()
	secrets, _, err := banzaiCLI.Client().SecretsApi.GetSecrets(
		context.Background(),
		orgID,
		&pipeline.GetSecretsOpts{
			Type_: optional.NewString(secretType),
		},
	)
	if err != nil {
		return "", errors.WrapIfWithDetails(err, "failed to get secret(s)", "secretType", secretType)
	}

	// hide hidden secrets
	var finalSecrets []pipeline.SecretItem
	for _, s := range secrets {
		if !isSecretHidden(s) {
			finalSecrets = append(finalSecrets, s)
		}
	}

	if len(finalSecrets) == 0 {
		// TODO (colin): add option to create new secret
		return "", nil
	}

	var secretName string
	var secretLen = len(finalSecrets)
	var secretIds = make(map[string]string, secretLen)

	secretOptions := make([]string, secretLen)

	for i, s := range finalSecrets {
		var idx = i
		secretOptions[idx] = s.Name
		secretIds[s.Name] = s.Id
	}

	if err := input.DoQuestions([]input.QuestionMaker{input.QuestionSelect{
		QuestionInput: input.QuestionInput{
			QuestionBase: input.QuestionBase{
				Message: "Provider secret:",
			},
			Output: &secretName,
		},
		Options: secretOptions,
	}}); err != nil {
		return "", errors.WrapIf(err, "error during getting secret")
	}

	return secretIds[secretName], nil
}

func isSecretHidden(s pipeline.SecretItem) bool {
	for _, t := range s.Tags {
		if t == "banzai:hidden" {
			return true
		}
	}
	return false
}
