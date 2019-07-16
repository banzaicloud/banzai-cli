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

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/utils"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
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
		Use:   "import",
		Short: "Import an existing cluster (EXPERIMENTAL)",
		Long:  "This is an experimental feature. You can import an existing Kubernetes cluster into Pipeline. Some Pipeline features may not work as expected.",
		Example: `banzai cluster import --name myimportedcluster --kubeconfig=kube.conf
kubectl config view --minify --raw | banzai cluster import -n myimportedcluster`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return importCluster(banzaiCli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.name, "name", "n", "", "Name of the cluster")
	flags.StringVarP(&options.file, "kubeconfig", "", "", "Kubeconfig file (with embed client cert/key for the user entry)")

	return cmd
}

func importCluster(banzaiCli cli.Cli, options importOptions) error {
	log.Warnln("This is an EXPERIMENTAL feature.")
	log.Warnln("Some Pipeline features may not work as expected.")

	client := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()

	if banzaiCli.Interactive() {
		var err error
		options.kubeconfig, options.name, err = buildInteractiveImportRequest(banzaiCli, options, orgId)
		if err != nil {
			return err
		}
		options.kubeconfig = base64.StdEncoding.EncodeToString([]byte(options.kubeconfig))
	} else {
		filename, raw, err := utils.ReadFileOrStdin(options.file)
		if err != nil {
			return emperror.WrapWith(err, "failed to read", "filename", filename)
		}
		options.kubeconfig = base64.StdEncoding.EncodeToString(raw)
	}

	// Create kubernetes secret
	req := pipeline.CreateSecretRequest{}
	req.Name = options.name
	req.Type = "kubernetes"
	req.Values = map[string]interface{}{
		"K8Sconfig": options.kubeconfig,
	}

	// Validate secret create request
	if err := validateSecretCreateRequest(req); err != nil {
		return emperror.Wrap(err, "failed to create cluster request")
	}

	req.Name += "-kubeconfig"

	if bytes, err := json.MarshalIndent(req, "", "  "); err != nil {
		log.Errorf("failed to marshal request: %v", err)
		log.Debugf("Request: %#v", req)
	} else {
		log.Debugf("The current state of the request:\n\n%s\n", bytes)
	}

	secret, resp, err := client.SecretsApi.AddSecrets(context.Background(), orgId, req, nil)
	if err != nil {
		// Secret already exists
		if resp != nil && resp.StatusCode == http.StatusConflict {
			return emperror.WrapWith(err, "secret with this name is already created", "secret", req.Name)
		}
		// Generic error
		if oerr, ok := err.(pipeline.GenericOpenAPIError); ok {
			log.Errorf("secret creation error: %s", oerr.Body())
		}

		return emperror.Wrap(err, "failed to create Kubernetes secret")
	}

	log.Debugf("Kubernetes config secret created with id: %s\n", secret.Id)

	// Create cluster with kubernetes secret
	body := map[string]interface{}{
		"name":     options.name,
		"secretId": secret.Id,
		"cloud":    "kubernetes",
		"properties": map[string]interface{}{
			"kubernetes": make(map[string]interface{}, 0),
		},
	}

	if bytes, err := json.MarshalIndent(body, "", "  "); err != nil {
		log.Errorf("failed to marshal request: %v", err)
		log.Debugf("Request: %#v", req)
	} else {
		log.Debugf("The current state of the request:\n\n%s\n", bytes)
	}

	cluster, _, err := client.ClustersApi.CreateCluster(context.Background(), orgId, body)
	if err != nil {
		// Generic error
		if oerr, ok := err.(pipeline.GenericOpenAPIError); ok {
			log.Errorf("secret creation error: %s", oerr.Body())
		}

		return emperror.Wrap(err, "failed to import Kubernetes cluster")
	}

	log.Debugf("Kubernetes config secret created with id: %d\n", cluster.Id)

	return nil
}

func buildInteractiveImportRequest(_ cli.Cli, options importOptions, _ int32) (kubeconfig, clusterName string, err error) {
	var fileName = options.file
	if fileName != "" {
		filename, raw, err := utils.ReadFileOrStdin(fileName)
		if err != nil {
			return "", "", emperror.WrapWith(err, "failed to read", "filename", filename)
		}
		kubeconfig = string(raw)
	}

	if kubeconfig == "" {
		_ = survey.AskOne(&survey.Editor{Message: "kubeconfig:", Default: ""}, &kubeconfig, nil)
	}

	name := fmt.Sprintf("%s%d", os.Getenv("USER"), os.Getpid())
	_ = survey.AskOne(&survey.Input{Message: "Cluster name:", Default: name}, &name, nil)
	clusterName = name

	return kubeconfig, clusterName, nil
}

func validateSecretCreateRequest(req pipeline.CreateSecretRequest) error {
	if req.Name == "" {
		return errors.New("cluster name must be specified")
	}
	if cfg, ok := req.Values["K8Sconfig"]; !ok || cfg == "" {
		return errors.New("kubernetes config must not be empty")
	}

	return nil
}
