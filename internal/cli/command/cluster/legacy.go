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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/banzaicloud/pipeline/client"
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
	Aliases: []string{"l"},
	Short:   "List clusters",
	Run:     ClusterList,
}

func ClusterList(cmd *cobra.Command, args []string) {
	pipeline := InitPipeline()
	orgId := GetOrgId(true)
	clusters, _, err := pipeline.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		logAPIError("list clusters", err, orgId)
		log.Fatalf("could not list clusters: %v", err)
	}
	Out(clusters, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt"})
}

var clusterGetCmd = &cobra.Command{
	Use:     "get NAME",
	Aliases: []string{"g"},
	Short:   "Get cluster details",
	Run:     ClusterGet,
	Args:    cobra.ExactArgs(1),
}

func ClusterGet(cmd *cobra.Command, args []string) {
	pipeline := InitPipeline()
	orgId := GetOrgId(true)
	clusters, _, err := pipeline.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		logAPIError("list clusters", err, orgId)
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
	cluster, _, err := pipeline.ClustersApi.GetCluster(context.Background(), orgId, id)
	if err != nil {
		logAPIError("get cluster", err, id)
		log.Fatalf("could not get cluster: %v", err)
	}
	type details struct {
		client.GetClusterStatusResponse
	}
	detailed := details{GetClusterStatusResponse: cluster}

	Out1(detailed, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt", "StatusMessage"})
}

var clusterDeleteCmd = &cobra.Command{
	Use:     "delete NAME",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a cluster",
	Run:     ClusterDelete,
	Args:    cobra.ExactArgs(1),
}

func ClusterDelete(cmd *cobra.Command, args []string) {
	pipeline := InitPipeline()
	orgId := GetOrgId(true)
	clusters, _, err := pipeline.ClustersApi.ListClusters(context.Background(), orgId)
	if err != nil {
		logAPIError("list clusters", err, orgId)
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

	if isInteractive() {
		if cluster, _, err := pipeline.ClustersApi.GetCluster(context.Background(), orgId, id); err != nil {
			logAPIError("get cluster", err, id)
		} else {
			Out1(cluster, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt", "StatusMessage"})
		}
		confirmed := false
		survey.AskOne(&survey.Confirm{Message: "Do you want to DELETE the cluster?"}, &confirmed, nil)
		if !confirmed {
			log.Fatal("deletion cancelled")
		}
	}
	if cluster, _, err := pipeline.ClustersApi.DeleteCluster(context.Background(), orgId, id, nil); err != nil {
		logAPIError("get cluster", err, id)
	} else {
		log.Printf("Deleting cluster %v", cluster)
	}
	if cluster, _, err := pipeline.ClustersApi.GetCluster(context.Background(), orgId, id); err != nil {
		logAPIError("get cluster", err, id)
	} else {
		Out1(cluster, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt", "StatusMessage"})
	}
}

var clusterShellCmd = &cobra.Command{
	Use:     "shell [command]",
	Aliases: []string{"sh", "k"},
	Short:   "Start a shell or run a command with the cluster configured as kubectl context",
	Run:     ClusterShell,
}

func ClusterShell(cmd *cobra.Command, args []string) {
	pipeline := InitPipeline()
	orgId := GetOrgId(true)
	id := GetClusterId(orgId, true)
	if id == 0 {
		log.Fatalf("no cluster selected")
	}

	cluster, _, err := pipeline.ClustersApi.GetCluster(context.Background(), orgId, id)
	if err != nil {
		log.Fatalf("could not get cluster details: %v", err)
	}

	config, _, err := pipeline.ClustersApi.GetClusterConfig(context.Background(), orgId, id)
	if err != nil {
		log.Fatalf("could not get cluster config: %v", err)
	}

	tmpfile, err := ioutil.TempFile("", "kubeconfig") // mode is 0600 by default
	if err != nil {
		log.Fatalf("could not write temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(config.Data)); err != nil {
		log.Fatalf("could not write temporary file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatalf("could not close temporary file: %v", err)
	}

	// customize shell prompt
	os.Setenv("PS1", fmt.Sprintf("[%s]$ ", chalk.Bold.TextStyle(cluster.Name)))
	// export the temporary config file's name for k8s commands
	os.Setenv("KUBECONFIG", tmpfile.Name())

	var commandArgs []string
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	if len(args) == 0 {
		// run interactive shell
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

	log.Printf("Running %v %v", shell, strings.Join(args, " "))
	c := exec.Command(shell, commandArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		log.Errorf("Failed to run command: %v", err)
	}
	log.Printf("Command exited successfully")
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
	clusterShellCmd.PersistentFlags().StringVar(&clusterOptions.Name, "cluster-name", "", "cluster name")
}
