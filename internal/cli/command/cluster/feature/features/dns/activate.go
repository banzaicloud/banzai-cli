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
)

func NewActivateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := activateOptions{}

	cmd := &cobra.Command{
		Use:     "activate",
		Aliases: []string{"add", "enable", "install", "on"},
		Short:   "Activate the DNS feature of a cluster",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runActivate(banzaiCli, options, args)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "activate DNS cluster feature for")

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Feature specification file")

	return cmd
}

type activateOptions struct {
	clustercontext.Context
	filePath string
}

func runActivate(banzaiCli cli.Cli, options activateOptions, _ []string) error {
	var req pipeline.ActivateClusterFeatureRequest
	if options.filePath == "" && banzaiCli.Interactive() {
		if err := buildActivateReqInteractively(banzaiCli, options, &req); err != nil {
			return emperror.Wrap(err, "failed to build activate request interactively")
		}
	} else {
		if err := readActivateReqFromFileOrStdin(options.filePath, &req); err != nil {
			return emperror.Wrap(err, "failed to read DNS cluster feature specification")
		}
	}

	orgId := banzaiCli.Context().OrganizationID()
	clusterId := options.ClusterID()
	resp, err := banzaiCli.Client().ClusterFeaturesApi.ActivateClusterFeature(context.Background(), orgId, clusterId, featureName, req)
	if err != nil {
		cli.LogAPIError("activate DNS cluster feature", err, resp.Request)
		log.Fatalf("could not activate DNS cluster feature: %v", err)
	}

	return nil
}

func readActivateReqFromFileOrStdin(filePath string, req *pipeline.ActivateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return emperror.WrapWith(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return emperror.Wrap(err, "failed to unmarshal input")
	}

	return nil
}

func buildActivateReqInteractively(_ cli.Cli, _ activateOptions, _ *pipeline.ActivateClusterFeatureRequest) error {
	panic("implement me")
	return nil
}
