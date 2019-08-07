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

package dns

import (
	"context"
	"encoding/json"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

func NewUpdateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := updateOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"change", "modify", "set"},
		Short:   "Update the DNS feature of a cluster",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runUpdate(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "update DNS cluster feature for")

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Feature specification file")

	return cmd
}

type updateOptions struct {
	clustercontext.Context
	filePath string
}

func runUpdate(banzaiCli cli.Cli, options updateOptions, _ []string) error {
	var req pipeline.UpdateClusterFeatureRequest
	if options.filePath == "" && banzaiCli.Interactive() {
		if err := buildUpdateReqInteractively(banzaiCli, options, &req); err != nil {
			return emperror.Wrap(err, "failed to build update request interactively")
		}
	} else {
		if err := readUpdateReqFromFileOrStdin(options.filePath, &req); err != nil {
			return emperror.Wrap(err, "failed to read DNS cluster feature specification")
		}
	}

	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()
	resp, err := banzaiCli.Client().ClusterFeaturesApi.UpdateClusterFeature(context.Background(), orgId, clusterId, featureName, req)
	if err != nil {
		cli.LogAPIError("activate DNS cluster feature", err, resp.Request)
		log.Fatalf("could not activate DNS cluster feature: %v", err)
	}

	return nil
}

func readUpdateReqFromFileOrStdin(filePath string, req *pipeline.UpdateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return emperror.WrapWith(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return emperror.Wrap(err, "failed to unmarshal input")
	}

	return nil
}

func buildUpdateReqInteractively(_ cli.Cli, _ updateOptions, req *pipeline.UpdateClusterFeatureRequest) error {
	var edit bool
	if err := survey.AskOne(&survey.Confirm{Message: "Do you want to edit the cluster feature update request in your text editor?"}, &edit, nil); err != nil {
		return emperror.Wrap(err, "failure during survey")
	}
	if !edit {
		return nil
	}

	content, err := json.MarshalIndent(*req, "", "  ")
	if err != nil {
		return emperror.Wrap(err, "failed to marshal request to JSON")
	}
	if err := survey.AskOne(&survey.Editor{Default: string(content), HideDefault: true, AppendDefault: true}, &content, validateActivateClusterFeatureRequest); err != nil {
		return emperror.Wrap(err, "failure during survey")
	}
	if err := json.Unmarshal(content, req); err != nil {
		return emperror.Wrap(err, "failed to unmarshal JSON as request")
	}

	return nil
}

func validateUpdateClusterFeatureRequest(req interface{}) error {
	var request pipeline.UpdateClusterFeatureRequest
	if err := json.Unmarshal([]byte(req.(string)), &request); err != nil {
		return emperror.Wrap(err, "request is not valid JSON")
	}

	return validateSpec(request.Spec)
}
