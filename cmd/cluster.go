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

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/antihax/optional"
	"github.com/banzaicloud/pipeline/client"
	yaml "github.com/ghodss/yaml"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ttacon/chalk"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

const clusterIdKey = "cluster.id"

var clusterOptions struct {
	Name string
}

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:     "cluster",
	Aliases: []string{"clusters", "c"},
	Short:   "Handle clusters",
	Run:     ClusterList,
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
	clusters, _, err := pipeline.ClustersApi.ListClusters(ctx, orgId)
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
	clusters, _, err := pipeline.ClustersApi.ListClusters(ctx, orgId)
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
	cluster, _, err := pipeline.ClustersApi.GetCluster(ctx, orgId, id)
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
	clusters, _, err := pipeline.ClustersApi.ListClusters(ctx, orgId)
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
		if cluster, _, err := pipeline.ClustersApi.GetCluster(ctx, orgId, id); err != nil {
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
	if cluster, _, err := pipeline.ClustersApi.DeleteCluster(ctx, orgId, id, nil); err != nil {
		logAPIError("get cluster", err, id)
	} else {
		log.Printf("Deleting cluster %v", cluster)
	}
	if cluster, _, err := pipeline.ClustersApi.GetCluster(ctx, orgId, id); err != nil {
		logAPIError("get cluster", err, id)
	} else {
		Out1(cluster, []string{"Id", "Name", "Distribution", "Status", "CreatorName", "CreatedAt", "StatusMessage"})
	}
}

var clusterCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c"},
	Short:   "create cluster based on json stdin or interactive session",
	Run:     ClusterCreate,
}

func unmarshal(raw []byte, data interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err == nil {
		return nil
	}

	// if can't decode as json, try to convert it from yaml first
	// use this method to prevent unmarshalling directly with yaml, for example to map[interface{}]interface{}
	converted, err := yaml.YAMLToJSON(raw)
	if err != nil {
		return emperror.Wrap(err, "unmarshal")
	}

	decoder = json.NewDecoder(bytes.NewReader(converted))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(data); err != nil {
		return emperror.Wrap(err, "unmarshal")
	}
	return nil
}

func ClusterCreate(cmd *cobra.Command, args []string) {
	pipeline := InitPipeline()
	orgId := GetOrgId(true)
	out := client.CreateClusterRequest{}
	if isInteractive() {
		content := ""
		for {
			fileName := ""
			survey.AskOne(&survey.Input{Message: "Load a JSON or YAML file:",
				Default: "skip",
				Help:    "Give either a relative or an absolute path to a file containing a JSON or YAML Cluster creation request. Leave empty to cancel."}, &fileName, nil)
			if fileName == "skip" || fileName == "" {
				break
			}
			if raw, err := ioutil.ReadFile(fileName); err != nil {
				log.Errorf("failed to read file %q: %v", fileName, err)
				continue
			} else {
				if err := unmarshal(raw, &out); err != nil {
					log.Fatalf("failed to parse CreateClusterRequest: %v", err)
				}
				break
			}
		}
		if out.Properties == nil || len(out.Properties) == 0 {
			providers := map[string]struct {
				cloud    string
				property interface{}
			}{
				"acsk": {cloud: "alibaba", property: new(client.CreateAkcsPropertiesAkcs)},
				"aks":  {cloud: "azure", property: new(client.CreateAksPropertiesAks)},
				"eks":  {cloud: "amazon", property: new(client.CreateEksPropertiesEks)},
				"gke":  {cloud: "google", property: new(client.CreateEksPropertiesEks)},
				"oke":  {cloud: "oracle", property: map[string]interface{}{}},
			}
			providerNames := make([]string, 0, len(providers))
			for provider := range providers {
				providerNames = append(providerNames, provider)
			}
			providerName := ""
			survey.AskOne(&survey.Select{Message: "Provider:", Help: "Select the provider to use", Options: providerNames}, &providerName, nil)
			if provider, ok := providers[providerName]; ok {
				out.Properties = map[string]interface{}{providerName: provider.property}
				out.Cloud = provider.cloud
			}
		}
		if out.SecretId == "" && out.SecretName == "" {
			secrets, _, err := pipeline.SecretsApi.GetSecrets(ctx, orgId, &client.GetSecretsOpts{Type_: optional.NewString(out.Cloud)})
			if err != nil {
				log.Errorf("could not list secrets: %v", err)
			} else {
				secretNames := make([]string, len(secrets))
				for i, secret := range secrets {
					secretNames[i] = secret.Name
				}
				survey.AskOne(&survey.Select{Message: "Secret:", Help: "Select the secret to use for creating cloud resources", Options: secretNames}, &out.SecretName, nil)
			}
		}
		if out.Name == "" {
			name := fmt.Sprintf("%s%s%d", os.Getenv("USER"), out.Cloud, os.Getpid())
			survey.AskOne(&survey.Input{Message: "Cluster name:", Default: name}, &out.Name, nil)
		}

		for {
			if bytes, err := json.MarshalIndent(out, "", "  "); err != nil {
				log.Errorf("failed to marshal request: %v", err)
				log.Debugf("Request: %#v", out)
			} else {
				content = string(bytes)
				fmt.Fprintf(os.Stderr, "The current state of the request:\n\n%s\n", content)
			}

			open := false
			survey.AskOne(&survey.Confirm{Message: "Do you want to edit the cluster request in your text editor?"}, &open, nil)
			if !open {
				break
			}
			///fmt.Printf("BEFORE>>>\n%v<<<\n", content)
			survey.AskOne(&survey.Editor{Message: "Create cluster request:", Default: content, HideDefault: true, AppendDefault: true}, &content, validateClusterCreateRequest)
			///fmt.Printf("AFTER>>>\n%v<<<\n", content)
			if err := json.Unmarshal([]byte(content), &out); err != nil {
				log.Errorf("can't parse request: %v", err)
			}
		}
		create := false
		survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Do you want to CREATE the cluster %q now?", out.Name)}, &create, nil)
		if !create {
			log.Fatal("cluster creation cancelled")
		}
	} else {
		// non-tty: read stdin
		raw, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("failed to read stdin: %v", err)
		}
		if err := unmarshal(raw, &out); err != nil {
			log.Fatalf("failed to parse CreateClusterRequest: %v", err)
		}
	}
	log.Debugf("create request: %#v", out)
	cluster, _, err := pipeline.ClustersApi.CreateCluster(ctx, orgId, out)
	if err != nil {
		logAPIError("create cluster", err, out)
		log.Fatalf("failed to create cluster: %v", err)
	}
	log.Info("cluster is being created")
	log.Infof("you can check its status with the command `banzai cluster get %q`", out.Name)
	Out1(cluster, []string{"Id", "Name", "Status"})
}

func validateClusterCreateRequest(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return errors.New("value is not a string")
	}
	decoder := json.NewDecoder(strings.NewReader(str))
	decoder.DisallowUnknownFields()
	return emperror.Wrap(decoder.Decode(&client.CreateClusterRequest{}), "not a valid JSON request")
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

	cluster, _, err := pipeline.ClustersApi.GetCluster(ctx, orgId, id)
	if err != nil {
		log.Fatalf("could not get cluster details: %v", err)
	}

	config, _, err := pipeline.ClustersApi.GetClusterConfig(ctx, orgId, id)
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
	clusters, _, err := pipeline.ClustersApi.ListClusters(ctx, org)
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
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterGetCmd)
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterShellCmd)
	clusterCmd.AddCommand(clusterDeleteCmd)

	clusterShellCmd.PersistentFlags().Int32("cluster", 0, "cluster id")
	viper.BindPFlag(clusterIdKey, clusterShellCmd.PersistentFlags().Lookup("cluster"))
	clusterShellCmd.PersistentFlags().StringVar(&clusterOptions.Name, "cluster-name", "", "cluster name")
}
