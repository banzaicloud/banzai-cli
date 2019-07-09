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

package cluster

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	pkgPipeline "github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type importOptions struct {
	name       string
	file       string
	kubeconfig string
}

// NewInstallCommand returns a cobra command for `install` subcommands.
func NewImportCommand(banzaiCli cli.Cli) *cobra.Command {
	options := importOptions{}

	cmd := &cobra.Command{
		Use:     "import",
		Aliases: []string{"imp"},
		Short:   "Manage cluster imports",
		RunE: func(cmd *cobra.Command, args []string) error {
			return importCluster(banzaiCli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.name, "name", "n", "", "Name of the cluster")
	flags.StringVarP(&options.file, "file", "f", "", "Cluster descriptor file")

	return cmd
}

func importCluster(banzaiCli cli.Cli, options importOptions) error {
	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()

	if banzaiCli.Interactive() {
		err := buildInteractiveImportRequest(banzaiCli, options, orgId)
		if err != nil {
			return err
		}
	} else {
		filename, raw, err := utils.ReadFileOrStdin(options.file)
		if err != nil {
			return emperror.WrapWith(err, fmt.Sprintf("failed to read %q", filename), "filename", filename)
		}

		options.kubeconfig = base64.StdEncoding.EncodeToString(raw)
	}

	// Create kubernetes secret
	req := pkgPipeline.CreateSecretRequest{}
	req.Name = options.name
	req.Type = "kubernetes"
	req.Values = make(map[string]interface{}, 1)
	req.Values["K8Sconfig"] = options.kubeconfig

	// Validate secret create request
	if err := validateSecretCreateRequest(req); err != nil {
		return emperror.Wrap(err, "failed to create cluster request")
	}

	req.Name += "-kubeconfig"

	if bytes, err := json.MarshalIndent(req, "", "  "); err != nil {
		log.Errorf("failed to marshal request: %v", err)
		log.Debugf("Request: %#v", req)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "The current state of the request:\n\n%s\n", bytes)
	}

	secret, resp, err := pipeline.SecretsApi.AddSecrets(context.Background(), orgId, req, nil)
	if err != nil {
		// Secret already exists
		if resp != nil && resp.StatusCode == http.StatusConflict {
			return emperror.WrapWith(err, "secret with this name is already created", "secret", req.Name)
		}
		// Generic error
		if oerr, ok := err.(pkgPipeline.GenericOpenAPIError); ok {
			log.Errorf("secret creation error: %s", oerr.Body())
		}

		return emperror.Wrap(err, "failed to create Kubernetes secret")
	}

	log.Debugf("Kubernetes config secret created with id: %s\n", secret.Id)

	// Create cluster with kubernetes secret
	body := make(map[string]interface{}, 4)
	body["name"] = options.name
	body["secretId"] = secret.Id
	body["cloud"] = "kubernetes"
	prop := make(map[string]interface{}, 1)
	prop["kubernetes"] = make(map[string]interface{}, 0)
	body["properties"] = prop

	if bytes, err := json.MarshalIndent(body, "", "  "); err != nil {
		log.Errorf("failed to marshal request: %v", err)
		log.Debugf("Request: %#v", req)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "The current state of the request:\n\n%s\n", bytes)
	}

	cluster, _, err := pipeline.ClustersApi.CreateCluster(context.Background(), orgId, body)
	if err != nil {
		// Generic error
		if oerr, ok := err.(pkgPipeline.GenericOpenAPIError); ok {
			log.Errorf("secret creation error: %s", oerr.Body())
		}

		return emperror.Wrap(err, "failed to import Kubernetes cluster")
	}

	log.Debugf("Kubernetes config secret created with id: %d\n", cluster.Id)

	return nil
}

func buildInteractiveImportRequest(banzaiCli cli.Cli, options importOptions, orgId int32) error {
	_ = banzaiCli
	_ = options
	_ = orgId
	// TODO: implement interactive version
	return errors.New("use --no-interactive")
}

func validateSecretCreateRequest(req pkgPipeline.CreateSecretRequest) error {
	if req.Name == "" {
		return errors.New("cluster name must be specified")
	}
	if cfg, ok := req.Values["K8Sconfig"]; !ok || cfg == "" {
		return errors.New("kubernetes config must not be empty")
	}

	return nil
}
