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

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"emperror.dev/errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/banzaicloud/banzai-cli/.gen/pipeline"
	"github.com/banzaicloud/banzai-cli/internal/cli"
	clustercontext "github.com/banzaicloud/banzai-cli/internal/cli/command/cluster/context"
	"github.com/banzaicloud/banzai-cli/pkg/sshconnector"
)

type nodeSSHOptions struct {
	clustercontext.Context

	nodeName        string
	directConnect   bool
	podConnect      bool
	punchThrough    bool
	username        string
	namespace       string
	useNodeAffinity bool
	useInternalIP   bool
	useExternalIP   bool
	sshPort         int
}

const (
	SSHSecretType  = "ssh"
	InternalIPType = "InternalIP"
	ExternalIPType = "ExternalIP"
)

func NewSSHToNodeCommand(banzaiCli cli.Cli) *cobra.Command {
	o := nodeSSHOptions{
		sshPort:         22,
		useNodeAffinity: false,
		namespace:       "pipeline-system",
	}

	cmd := &cobra.Command{
		Use:     "ssh [NODE_NAME]",
		Aliases: []string{"c", "connect"},
		Short:   "Connect to node with SSH",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.punchThrough {
				o.podConnect = true
				o.useInternalIP = true
			}
			if !o.directConnect && !o.podConnect {
				o.directConnect = true
			}
			if o.directConnect && o.podConnect {
				return fmt.Errorf("--direct-connect and --pod-connect are mutually exclusive")
			}
			if !o.useInternalIP && !o.useExternalIP {
				o.useExternalIP = true
			}
			if o.useInternalIP && o.useExternalIP {
				return fmt.Errorf("--use-internal-ip and --use-external-ip are mutually exclusive")
			}
			return runzSSHToNode(banzaiCli, o, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&o.nodeName, "node-name", o.nodeName, "Name of Kubernetes node to connect to")
	flags.StringVar(&o.username, "username", o.username, "Username to use for the SSH connection")
	flags.BoolVarP(&o.punchThrough, "punch-through", "p", o.punchThrough, "Shorthand for --pod-connect --use-internal-ip")
	flags.BoolVar(&o.directConnect, "direct-connect", o.directConnect, "Use direct connection to the node internal or external IP (default)")
	flags.BoolVar(&o.podConnect, "pod-connect", o.podConnect, "Create a pod on one of the nodes and connect to a node through that pod")
	flags.StringVar(&o.namespace, "namespace", o.namespace, "Namespace for the pod when using --pod-connect")
	flags.BoolVar(&o.useNodeAffinity, "use-node-affinity", o.useNodeAffinity, "Whether to use node affinity for pod scheduling when using --pod-connect")
	flags.BoolVar(&o.useInternalIP, "use-internal-ip", o.useInternalIP, "Use internal IP of the node to connect")
	flags.BoolVar(&o.useExternalIP, "use-external-ip", o.useExternalIP, "Use external IP of the node to connect (default)")
	flags.IntVar(&o.sshPort, "ssh-port", o.sshPort, "SSH port of the node to connect")

	o.Context = clustercontext.NewClusterContext(cmd, banzaiCli, "get")

	return cmd
}

func runzSSHToNode(banzaiCli cli.Cli, options nodeSSHOptions, args []string) error {
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

	var nodeName string
	if len(args) > 0 {
		nodeName = args[0]
	}
	if nodeName == "" && options.nodeName != "" {
		nodeName = options.nodeName
	}

	if nodeName == "" && !banzaiCli.Interactive() {
		return errors.New("no node is selected; use the --node-name option or add node name as an argument")
	}

	node, err := getOrSelectNode(client, orgID, clusterID, nodeName)
	if err != nil {
		return err
	}

	secret, err := getSSHSecretForCluster(client, orgID, clusterID)
	if err != nil {
		return err
	}

	tmpfile, err := createTempFile("", "sshkey", secret.Values["private_key_data"].(string))
	if err != nil {
		return err
	}
	defer func() {
		os.Remove(tmpfile.Name())
	}()

	var ipAddress string
	for _, address := range node.Status.Addresses {
		if address.Type == InternalIPType && options.useInternalIP {
			ipAddress = address.Address
		}
		if address.Type == ExternalIPType && options.useExternalIP {
			ipAddress = address.Address
		}
	}

	cluster, _, err := client.ClustersApi.GetCluster(context.Background(), orgID, clusterID)
	if err != nil {
		return err
	}

	var username string
	if options.username != "" {
		username = options.username
	} else {
		switch cluster.Distribution {
		case "oke":
			username = "opc"
		case "aks":
			username = "aks-user"
		case "eks":
			username = "ec2-user"
		}

		if !banzaiCli.Interactive() && username == "" {
			return errors.New("can't determine username to use for the connection (you can specify it with an option like --username=ubuntu)")
		}

		if banzaiCli.Interactive() {
			err = survey.AskOne(&survey.Input{
				Message: "Username:",
				Default: username,
				Help:    "The username to use for the SSH connection, for example ubuntu, centos, root, or ec2-user.",
			}, &username, survey.WithValidator(survey.Required))
			if err != nil {
				return errors.WrapIf(err, "failed to select username")
			}
		}
	}

	if options.directConnect {
		connector := sshconnector.NewDirectSSHConnector()
		connector.Connect(ipAddress, options.sshPort, username, tmpfile.Name())
	} else {
		kubeconfigfile, err := getClusterKubeconfigFile(client, orgID, clusterID)
		if err != nil {
			return err
		}
		var opts []sshconnector.PodSSHConnectorOption
		if options.namespace != "" {
			opts = append(opts, sshconnector.NamespaceOption(options.namespace))
		}
		if options.useNodeAffinity {
			opts = append(opts, sshconnector.NodeNameOption(node.Metadata.Name))
		}
		connector, err := sshconnector.NewPodSSHConnector(kubeconfigfile.Name(), opts...)
		if err != nil {
			return err
		}
		defer connector.Cleanup()

		err = connector.Connect(ipAddress, options.sshPort, username, tmpfile.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func getOrSelectNode(client *pipeline.APIClient, orgID, clusterID int32, nodeName string) (*pipeline.NodeItem, error) {
	var node pipeline.NodeItem

	nodes, _, err := client.ClustersApi.ListNodes(context.Background(), orgID, clusterID)
	if err != nil {
		return nil, errors.WrapIf(convertError(err), "could not list nodes")
	}

	if nodeName == "" {
		nodeNames := make(map[string]string)
		var selectedNodeName string
		nodeOptions := make([]string, 0)
		for _, n := range nodes.Items {
			key := fmt.Sprintf("%s (%s)", n.Metadata.Name, n.Metadata.Labels["nodepool.banzaicloud.io/name"])
			nodeNames[key] = n.Metadata.Name
			nodeOptions = append(nodeOptions, key)
		}

		err = survey.AskOne(&survey.Select{Message: "Node:", Options: nodeOptions}, &selectedNodeName, survey.WithValidator(survey.Required))
		if err != nil {
			return nil, errors.WrapIf(err, "failed to select a node")
		}

		nodeName = nodeNames[selectedNodeName]
	}

	for _, _node := range nodes.Items {
		if _node.Metadata.Name == nodeName {
			node = _node
		}
	}

	if node.Metadata.Name == "" {
		return nil, fmt.Errorf("could not find node: %s", nodeName)
	}

	return &node, nil
}

func createTempFile(dir, pattern, content string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile(dir, pattern) // mode is 0600 by default
	if err != nil {
		return nil, errors.WrapIf(err, "could not write temporary file")
	}

	if _, err = tmpfile.Write([]byte(content)); err != nil {
		return nil, errors.WrapIf(err, "could not write temporary file")
	}

	if err = tmpfile.Close(); err != nil {
		return nil, errors.WrapIf(err, "could not close temporary file")
	}

	return tmpfile, nil
}

func getClusterKubeconfigFile(client *pipeline.APIClient, orgID, clusterID int32) (*os.File, error) {
	kubeconfig, _, clusterErr := client.ClustersApi.GetClusterConfig(context.Background(), orgID, clusterID)
	if clusterErr != nil {
		return nil, errors.WrapIf(clusterErr, "could not get cluster config")
	}

	kubetmpfile, err := createTempFile("", "kubeconfig", kubeconfig.Data)
	if err != nil {
		return nil, errors.WrapIf(err, "could not write temporary file")
	}

	return kubetmpfile, nil
}

func getSSHSecretForCluster(client *pipeline.APIClient, orgID int32, clusterID int32) (pipeline.SecretItem, error) {
	var secret pipeline.SecretItem

	secrets, _, err := client.ClustersApi.ListClusterSecrets(context.Background(), orgID, clusterID, nil)
	if err != nil {
		return secret, err
	}

	for _, _secret := range secrets {
		if _secret.Type == SSHSecretType {
			secret, _, err := client.SecretsApi.GetSecret(context.Background(), orgID, _secret.Id)
			return secret, err
		}
	}

	return secret, fmt.Errorf("could not find secret for cluster")
}

func convertError(err error) error {
	type Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if gerr, ok := errors.Cause(err).(pipeline.GenericOpenAPIError); ok {
		var pipelineError Error
		e := json.Unmarshal(gerr.Body(), &pipelineError)
		if e == nil {
			return errors.WithMessage(err, pipelineError.Message)
		}
	}

	return err
}
