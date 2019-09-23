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

package securityscan

import (
	"context"
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewActivateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := activateOptions{}

	cmd := &cobra.Command{
		Use:           "activate",
		Aliases:       []string{"add", "enable", "install", "on"},
		Short:         fmt.Sprintf("Activate the %s feature of a cluster", featureName),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runActivate(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, fmt.Sprintf("activate %s cluster feature for", featureName))

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Feature specification file")

	return cmd
}

type activateOptions struct {
	clustercontext.Context
	filePath string
}

func runActivate(banzaiCli cli.Cli, options activateOptions, args []string) error {

	if err := options.Init(args...); err != nil {
		return errors.Wrap(err, "failed to initialize options")
	}

	var req pipeline.ActivateClusterFeatureRequest
	if options.filePath == "" && banzaiCli.Interactive() {
		if err := buildActivateReqInteractively(banzaiCli, &req); err != nil {
			return errors.WrapIf(err, "failed to build activate request interactively")
		}
	} else {
		if err := readActivateReqFromFileOrStdin(options.filePath, &req); err != nil {
			return errors.WrapIff(err, "failed to read %s cluster feature specification", featureName)
		}
	}

	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()
	_, err := banzaiCli.Client().ClusterFeaturesApi.ActivateClusterFeature(context.Background(), orgId, clusterId, featureName, req)
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("activate %s cluster feature", featureName), err, req)
		log.Fatalf("could not activate %s cluster feature: %v", featureName, err)
	}

	log.Infof("feature %q started to activate", featureName)

	return nil
}

func readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func buildActivateReqInteractively(
	banzaiCli cli.Cli,
	req *pipeline.ActivateClusterFeatureRequest,
) error {

	var edit bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Do you want to edit the cluster feature activation request in your text editor?",
		},
		&edit,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if !edit {
		return nil
	}

	content, err := json.MarshalIndent(*req, "", "  ")
	if err != nil {
		return errors.WrapIf(err, "failed to marshal request to JSON")
	}
	var result string
	if err := survey.AskOne(
		&survey.Editor{
			Default:       string(content),
			HideDefault:   true,
			AppendDefault: true,
		},
		&result,
		survey.WithValidator(validateActivateClusterFeatureRequest),
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if err := json.Unmarshal([]byte(result), req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}

func validateActivateClusterFeatureRequest(req interface{}) error {
	var request pipeline.ActivateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return errors.WrapIf(err, "request is not valid JSON")
	}

	return nil
}

func buildScurityscanFeatureRequest(banzaiCli cli.Cli, req *pipeline.ActivateClusterFeatureRequest) error {

	return nil
}
