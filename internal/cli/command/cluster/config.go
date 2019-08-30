// Copyright © 2018 Banzai Cloud
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
	"fmt"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type configOptions struct {
	clustercontext.Context
	oidc bool
}

func NewConfigCommand(banzaiCli cli.Cli) *cobra.Command {
	options := configOptions{}
	cmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"conf"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfig(banzaiCli, options)
		},
		Short:   "Downloads a cluster's kubectl config",
		Long:    "You can either run the command without arguments to interactively select a cluster, and get an interactive shell, select the cluster with the --cluster-name flag, or specify the command to run.",
		Example: ``,
	}

	cmd.Flags().BoolVar(&options.oidc, "oidc", true, "Wrap the helm command with a version that downloads the matching version and creates a custom helm home")

	options.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "run a shell for")

	return cmd
}

func runConfig(banzaiCli cli.Cli, options configOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()
	if err := options.Init(); err != nil {
		return err
	}
	id := options.ClusterID()

	kubeconfigPath := fmt.Sprintf("%s/kubeconfig-%d-%d.yaml", banzaiCli.Home(), orgId, id)
	kubeconfigFile, err := os.OpenFile(kubeconfigPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open kube config file for writing")
	}

	interactive := banzaiCli.Interactive()

	if options.oidc {
		if !interactive {
			return errors.New("oidc config available only in interactive mode")
		}

		oidcSecret, _, err := pipeline.SecretsApi.GetSecret(ctx, orgId, fmt.Sprintf("cluster-%d-dex-client", id))
		if err != nil {
			return errors.WrapIf(err, "could not get oidc secret")
		}

		runOIDCServer(oidcSecret.Values["clientID"].(string), oidcSecret.Values["clientSecret"].(string))

	} else {

		retry, err := writeConfig(ctx, pipeline, orgId, id, kubeconfigFile)
		if err != nil {
			if !interactive || !retry {
				return errors.WrapIf(err, "writing kubeconfig")
			}

			go func() {
				for {
					retry, err := writeConfig(ctx, pipeline, orgId, id, kubeconfigFile)
					if err != nil {
						if !retry {
							log.Fatalf("%v", err)
						}
						log.Warningf("cluster config is still not available. retrying in 30 seconds")
					} else {
						log.Infof("cluster config successfully written")
						return
					}

					select {
					case <-time.After(30 * time.Second):
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	return nil
}

func runOIDCServer(clientID, clientSecret string) {}
