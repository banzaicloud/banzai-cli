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

package services

import (
	"context"
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
)

type updateOptions struct {
	clustercontext.Context
	filePath string
}

type UpdateManager interface {
	GetName() string
	ValidateRequest(interface{}) error
	BuildRequestInteractively(
		banzaiCli cli.Cli,
		updateServiceRequest *pipeline.UpdateClusterFeatureRequest,
		clusterCtx clustercontext.Context,
		cap map[string]interface{},
	) error
}

func UpdateCommandFactory(banzaiCLI cli.Cli, use string, manager UpdateManager, name string) *cobra.Command {
	options := updateOptions{}

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"change", "modify", "set"},
		Short:   fmt.Sprintf("Update the %s service of a cluster", name),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runUpdate(banzaiCLI, manager, options, args, use)
		},
	}

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCLI, fmt.Sprintf("update %s cluster service for", name))

	flags := cmd.Flags()
	flags.StringVarP(&options.filePath, "file", "f", "", "Service specification file")

	return cmd
}

func runUpdate(
	banzaiCLI cli.Cli,
	m UpdateManager,
	options updateOptions,
	args []string,
	use string,
) error {
	var cl = capLoader{cli: banzaiCLI}
	capabilities, err := cl.loadCapabilities(context.Background(), use)
	if err != nil {
		return errors.WrapIf(err, "error during loading capabilities")
	}

	if err := capabilities.isServiceEnabled(); err != nil {
		return errors.WrapIf(err, "failed to check service")
	}

	if err := options.Init(args...); err != nil {
		return errors.Wrap(err, "failed to initialize options")
	}

	orgID := banzaiCLI.Context().OrganizationID()
	clusterID := options.ClusterID()

	var request *pipeline.UpdateClusterFeatureRequest
	if options.filePath == "" && banzaiCLI.Interactive() {

		// get integratedservice details
		details, _, err := banzaiCLI.Client().ClusterFeaturesApi.ClusterFeatureDetails(context.Background(), orgID, clusterID, m.GetName())
		if err != nil {
			return errors.WrapIf(err, "failed to get service details")
		}

		request = &pipeline.UpdateClusterFeatureRequest{
			Spec: details.Spec,
		}

		if err := m.BuildRequestInteractively(banzaiCLI, request, options.Context, capabilities); err != nil {
			return errors.WrapIf(err, "failed to build update request interactively")
		}

		// show editor
		if err := showUpdateEditor(m, request); err != nil {
			return errors.WrapIf(err, "failed during showing editor")
		}

	} else {
		if err := readUpdateReqFromFileOrStdin(options.filePath, request); err != nil {
			return errors.WrapIf(err, fmt.Sprintf("failed to read %s cluster service specification", m.GetName()))
		}
	}

	resp, err := banzaiCLI.Client().ClusterFeaturesApi.UpdateClusterFeature(context.Background(), orgID, clusterID, m.GetName(), *request)
	if err != nil {
		cli.LogAPIError(fmt.Sprintf("update %s cluster service", m.GetName()), err, resp.Request)
		log.Fatalf("could not update %s cluster service: %v", m.GetName(), err)
	}

	log.Infof("service %q started to update", m.GetName())

	return nil
}

func readUpdateReqFromFileOrStdin(filePath string, req *pipeline.UpdateClusterFeatureRequest) error {
	filename, raw, err := utils.ReadFileOrStdin(filePath)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return errors.WrapIf(err, "failed to unmarshal input")
	}

	return nil
}

func showUpdateEditor(m UpdateManager, request *pipeline.UpdateClusterFeatureRequest) error {
	var edit bool
	if err := survey.AskOne(
		&survey.Confirm{
			Message: "Do you want to edit the cluster service update request in your text editor?",
		},
		&edit,
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if !edit {
		return nil
	}

	content, err := json.MarshalIndent(*request, "", "  ")
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
		survey.WithValidator(m.ValidateRequest),
	); err != nil {
		return errors.WrapIf(err, "failure during survey")
	}
	if err := json.Unmarshal([]byte(result), &request); err != nil {
		return errors.WrapIf(err, "failed to unmarshal JSON as request")
	}

	return nil
}
