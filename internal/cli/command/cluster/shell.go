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
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/goph/emperror"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

type shellOptions struct {
	Context
}

func NewShellCommand(banzaiCli cli.Cli) *cobra.Command {
	options := shellOptions{}
	cmd := &cobra.Command{
		Use:     "shell [command]",
		Aliases: []string{"sh"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShell(banzaiCli, options, args)
		},
		Short: "Start a shell or run a command with the cluster configured as kubectl context",
		Long: "The banzai CLI's cluster shell command starts your default shell, or runs your specified program on your local machine within the Kubernetes context of your cluster. " +
			"You can either run the command without arguments to interactively select a cluster, and get an interactive shell, select the cluster with the --cluster-name flag, or specify the command to run.",
		Example: `
			$ banzai cluster shell
			? Cluster: docs-example
			[docs-example]$ helm list
			...
			[docs-example]$ kubectl get nodes
			...
			[docs-example]$ exit
			INFO[0026] Command exited successfully

			$ banzai cluster shell --cluster-name docs-example kubectl get nodes
			INFO[0000] Running kubectl kubectl get nodes
			NAME                                    STATUS   ROLES    AGE   VERSION
			gke-docs-example-pool1-7a602b82-62w8    Ready    <none>   43m   v1.10.11-gke.1
			gke-docs-example-system-a16f163c-dvwj   Ready    <none>   43m   v1.10.11-gke.1
			INFO[0001] Command exited successfully`,
	}

	options.Context = NewClusterContext(cmd, banzaiCli, "run a shell for")

	return cmd
}

func writeConfig(ctx context.Context, client *pipeline.APIClient, orgId, id int32, tmpfile io.WriteCloser) (retry bool, err error) {
	config, response, clusterErr := client.ClustersApi.GetClusterConfig(ctx, orgId, id)
	if clusterErr != nil {
		retry = response.StatusCode == 400
		err = emperror.Wrap(clusterErr, "could not get cluster config")
		return
	}

	if _, err = tmpfile.Write([]byte(config.Data)); err != nil {
		err = emperror.Wrap(err, "could not write temporary file")
		return
	}

	if err = tmpfile.Close(); err != nil {
		err = emperror.Wrap(err, "could not close temporary file")
	}
	return
}

func runShell(banzaiCli cli.Cli, options shellOptions, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pipeline := banzaiCli.Client()
	orgId := banzaiCli.Context().OrganizationID()
	if err := options.Init(); err != nil {
		return err
	}
	id := options.ClusterID()

	tmpfile, err := ioutil.TempFile("", "kubeconfig") // mode is 0600 by default
	if err != nil {
		return emperror.Wrap(err, "could not write temporary file")
	}
	defer os.Remove(tmpfile.Name())

	var commandArgs []string
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	// if no args are specified, we start a[n interactive] shell, otherwise run the command from args
	interactive := len(args) == 0

	if interactive {
		switch path.Base(shell) {
		case "zsh": // zsh (at least my config) overrides PS1 :(
			fallthrough
		case "bash":
			commandArgs = []string{"-i"}
		}

	} else if len(args) == 1 {
		// let the shell split arg to words
		commandArgs = []string{"-c", args[0]}

	} else {
		// exec args as separate words
		shell = args[0]
		commandArgs = args[1:]
	}

	retry, err := writeConfig(ctx, pipeline, orgId, id, tmpfile)
	if err != nil {
		if !interactive || !retry {
			return emperror.Wrap(err, "writing kubeconfig")
		}

		go func() {
			for {
				retry, err := writeConfig(ctx, pipeline, orgId, id, tmpfile)
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

	org, _, err := pipeline.OrganizationsApi.GetOrg(ctx, orgId)
	if err != nil {
		return emperror.Wrap(err, "could not get organization")
	}

	log.Printf("Running %v %v", shell, strings.Join(args, " "))
	c := exec.Command(shell, commandArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	env := []string{
		// customize shell prompt
		fmt.Sprintf("PS1=[%s]$ ", chalk.Bold.TextStyle(options.ClusterName())),

		// export the temporary config file's name for k8s commands
		fmt.Sprintf("KUBECONFIG=%s", tmpfile.Name()),

		fmt.Sprintf("BANZAI_CURRENT_ORG_ID=%d", orgId),
		fmt.Sprintf("BANZAI_CURRENT_ORG_NAME=%s", org.Name),
		fmt.Sprintf("BANZAI_CURRENT_CLUSTER_ID=%d", id),
		fmt.Sprintf("BANZAI_CURRENT_CLUSTER_NAME=%s", options.ClusterName()),
	}

	log.Debugf("Environment: %s", strings.Join(env, " "))
	c.Env = append(os.Environ(), env...)

	if err := c.Run(); err != nil {
		wrapped := emperror.Wrap(err, "failed to run command")

		if err, ok := err.(interface{ ExitCode() int }); ok {
			log.Error(wrapped)
			os.Exit(err.ExitCode())
		}

		return wrapped
	}
	log.Printf("Command exited successfully")
	return nil
}
