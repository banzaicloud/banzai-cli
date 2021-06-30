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

package nodepool

import (
	"context"
	"encoding/json"
	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createOptions struct {
	clustercontext.Context

	file string

	name string
}

func NewCreateCommand(banzaiCli cli.Cli) *cobra.Command {
	options := createOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a node pool",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			return createNodePool(banzaiCli, options)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&options.file, "file", "f", "", "Node pool descriptor file")
	flags.StringVarP(&options.name, "name", "n", "", "Node pool name")

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "create")

	return cmd
}

func parseNodePoolCreateRequest(raw []byte) ([]pipeline.NodePool, error) {
	str := string(raw)
	jsonDecoder := json.NewDecoder(strings.NewReader(str))

	var rawRequest interface{}
	err := jsonDecoder.Decode(&rawRequest)
	if err != nil {
		return nil, errors.WrapIf(err, "invalid JSON request")
	}

	if _, isArray := rawRequest.([]interface{}); isArray {
		var request []pipeline.NodePool
		if err := utils.Unmarshal(raw, &request); err != nil {
			return nil, errors.WrapIf(err, "failed to unmarshal create node pool request")
		}
		return request, nil
	}

	var request pipeline.NodePool
	if err := utils.Unmarshal(raw, &request); err != nil {
		return nil, errors.WrapIf(err, "failed to unmarshal create node pool request")
	}
	return []pipeline.NodePool{request}, nil
}

func createNodePool(banzaiCli cli.Cli, options createOptions) error {
	client := banzaiCli.Client()
	orgID := banzaiCli.Context().OrganizationID()

	err := options.Init()
	if err != nil {
		return err
	}

	clusterID := options.ClusterID()
	if clusterID == 0 {
		return errors.New("no clusters found")
	}

	filename, raw, err := utils.ReadFileOrStdin(options.file)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to read", "filename", filename)
	}

	log.Debugf("%d bytes read", len(raw))

	request, err := parseNodePoolCreateRequest(raw)
	if err != nil {
		return errors.WrapIfWithDetails(err, "failed to parse create node pool request")
	}

	if options.name != "" {
		for _, req := range request {
			req.Name = options.name
		}
	}

	log.Debugf("create request: %#v", request)

	resp, err := client.ClustersApi.CreateNodePool(context.Background(), orgID, clusterID, request)
	if err != nil {
		cli.LogAPIError("create node pool", err, request)

		return errors.WrapIf(err, "failed to create node pool")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.NewWithDetails("node pool creation failed with http status code", "status_code", resp.StatusCode)

		cli.LogAPIError("create node pool", err, request)

		return err
	}

	return nil
}
