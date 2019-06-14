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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
	"gopkg.in/AlecAivazis/survey.v1"
)

const clusterIdKey = "cluster.id"

var clusterOptions struct {
	Name string
}

var clusterListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List clusters",
	Run:     ClusterList,
}

func ClusterList(cmd *cobra.Command, args []string) {
	pipeline := InitPipeline()
	orgId := GetOrgId(true)
	clusters, _, err := pipeline.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		cli.LogAPIError("list clusters", err, orgId)
		log.Fatalf("could not list clusters: %v", err)
	}
	Out(clusters, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt"})
}

var clusterGetCmd = &cobra.Command{
	Use:     "get NAME",
	Aliases: []string{"g", "show"},
	Short:   "Get cluster details",
	Run:     ClusterGet,
	Args:    cobra.ExactArgs(1),
}

func ClusterGet(cmd *cobra.Command, args []string) {
	client := InitPipeline()
	orgId := GetOrgId(true)
	clusters, _, err := client.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		cli.LogAPIError("list clusters", err, orgId)
		log.Fatalf("could not list clusters: %v", err)
	}
	var id int32
	for _, cluster := range clusters {
		if cluster.Name == args[0] || fmt.Sprintf("%d", cluster.Id) == args[0] {
			id = cluster.Id
			break
		}
	}
	if id == 0 {
		log.Fatalf("cluster %q could not be found", args[0])
	}
	cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgId, id)
	if err != nil {
		cli.LogAPIError("get cluster", err, id)
		log.Fatalf("could not get cluster: %v", err)
	}
	type details struct {
		pipeline.GetClusterStatusResponse
	}
	detailed := details{GetClusterStatusResponse: cluster}

	Out1(detailed, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt", "StatusMessage"})
}

var clusterShellCmd = &cobra.Command{
	RunE:    ClusterShell,
	Use:     "shell [command]",
	Aliases: []string{"sh"},
	Short:   "Start a shell or run a command with the cluster configured as kubectl context",
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

func ClusterShell(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pipeline := InitPipeline()
	orgId := GetOrgId(true)
	id := GetClusterId(orgId, true)
	if id == 0 {
		return errors.New("no cluster selected")
	}

	cluster, _, err := pipeline.ClustersApi.GetCluster(ctx, orgId, id)
	if err != nil {
		return emperror.Wrap(err, "could not get cluster details")
	}

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
		fmt.Sprintf("PS1=[%s]$ ", chalk.Bold.TextStyle(cluster.Name)),

		// export the temporary config file's name for k8s commands
		fmt.Sprintf("KUBECONFIG=%s", tmpfile.Name()),

		fmt.Sprintf("BANZAI_CURRENT_ORG_ID=%d", orgId),
		fmt.Sprintf("BANZAI_CURRENT_ORG_NAME=%s", org.Name),
		fmt.Sprintf("BANZAI_CURRENT_CLUSTER_ID=%d", id),
		fmt.Sprintf("BANZAI_CURRENT_CLUSTER_NAME=%s", cluster.Name),
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

func GetClusterId(org int32, ask bool) int32 {
	pipeline := InitPipeline()
	clusters, _, err := pipeline.ClustersApi.ListClusters(context.Background(), org)
	if err != nil {
		log.Fatalf("could not list clusters: %v", err)
	}

	if n := clusterOptions.Name; n != "" {
		for _, cluster := range clusters {
			if n == cluster.Name {
				return cluster.Id
			}
		}
	}

	id := viper.GetInt32(clusterIdKey)
	if id != 0 {
		return id
	}

	if ask && !isInteractive() {
		log.Fatal("No cluster is selected. Use the --cluster or --cluster-name switch, or set the cluster.id config value.")
	}
	clusterSlice := make([]string, len(clusters))
	for i, cluster := range clusters {
		clusterSlice[i] = cluster.Name
	}
	name := ""
	survey.AskOne(&survey.Select{Message: "Cluster:", Options: clusterSlice}, &name, survey.Required)
	for _, cluster := range clusters {
		if name == cluster.Name {
			return cluster.Id
		}
	}
	return 0
}

func init() {
	clusterShellCmd.PersistentFlags().Int32("cluster", 0, "cluster id")
	viper.BindPFlag(clusterIdKey, clusterShellCmd.PersistentFlags().Lookup("cluster"))
	viper.BindEnv(clusterIdKey, "BANZAI_CURRENT_CLUSTER_ID")
	clusterShellCmd.PersistentFlags().StringVar(&clusterOptions.Name, "cluster-name", "", "cluster name")
}
