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
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/input"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createOptions struct {
	clustercontext.Context

	filePath string
}

func newCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a manual backup",
		Long:    "Create a one-time manual backup.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			if err := options.Init(args...); err != nil {
				return errors.WrapIf(err, "failed to initialize options")
			}

			return runCreate(banzaiCli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Create backup specification file")

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "create")

	return cmd
}

func runCreate(banzaiCli cli.Cli, options createOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()
	clusterID := options.ClusterID()

	enabled, err := isCommandEnabledForCluster(client, orgID, clusterID)
	if err != nil {
		return errors.WrapIf(err, "error during checking command availability")
	}

	if !enabled {
		return NotAvailableError{}
	}

	var request pipeline.CreateBackupRequest
	if options.filePath == "" && banzaiCli.Interactive() {
		if request, err = buildCreateRequestInteractively(); err != nil {
			return errors.WrapIf(err, "failed to build create backup request interactively")
		}
	} else {
		if err = readCreateReqFromFileOrStdin(options.filePath, &request); err != nil {
			return errors.WrapIf(err, "failed to read create backup specification")
		}
	}

	_, _, err = client.ArkBackupsApi.CreateARKBackupOfACluster(context.Background(), orgID, clusterID, request)
	if err != nil {
		return errors.WrapIf(err, "failed to create backup")
	}

	log.Infof("Backup created for cluster [%d]", clusterID)

	return nil
}

func buildCreateRequestInteractively() (pipeline.CreateBackupRequest, error) {
	var name string
	var ttlLabel string

	var currentTime = time.Now()

	if err := input.DoQuestions([]input.QuestionMaker{
		input.QuestionInput{
			QuestionBase: input.QuestionBase{
				Message: "Name of the backup",
				Help:    fmt.Sprintf("Name of the backup, for example, `manual-backup-%s`", currentTime.Format("2006-01-02")),
			},
			Output: &name,
		},
		input.QuestionSelect{
			QuestionInput: input.QuestionInput{
				QuestionBase: input.QuestionBase{
					Message: "Keep backup for",
					Help:    "Retain backup for the specified period.",
				},
				DefaultValue: ttl1DayLabel,
				Output:       &ttlLabel,
			},
			Options: []string{ttl1DayLabel, ttl2DaysLabel, ttl1WeekLabel},
		},
	}); err != nil {
		return pipeline.CreateBackupRequest{}, errors.WrapIf(err, "error during getting create options")
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

	return pipeline.CreateBackupRequest{
		Name: name,
		Ttl:  selectedTTL,
	}, nil
}

func readCreateReqFromFileOrStdin(filePath string, req *pipeline.CreateBackupRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIfWithDetails(err,
			"failed to unmarshal input",
			"fileName", filename,
			"raw request", string(raw),
		)
	}

	return nil
}
