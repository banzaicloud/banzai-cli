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

package sshconnector

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	podNamePrefix    = "ssh-pod"
	kubectlCmd       = "kubectl"
	sshCmd           = "ssh"
	defaultNamespace = "pipeline-system"
	localPort        = 12389
)

type podSSHConnector struct {
	kubeconfig string
	namespace  string
	nodeName   string

	podName        string
	podCreated     bool
	podRunning     bool
	portforwardPID int

	shutdown bool

	logger log.FieldLogger
}

type PodSSHConnectorOption func(*podSSHConnector)

func NamespaceOption(namespace string) PodSSHConnectorOption {
	return func(opts *podSSHConnector) {
		opts.namespace = namespace
	}
}

func NodeNameOption(name string) PodSSHConnectorOption {
	return func(opts *podSSHConnector) {
		opts.nodeName = name
	}
}

func NewPodSSHConnector(kubeconfig string, opts ...PodSSHConnectorOption) (*podSSHConnector, error) {
	connector := &podSSHConnector{
		kubeconfig: kubeconfig,
		namespace:  defaultNamespace,
	}

	for _, opt := range opts {
		opt(connector)
	}

	err := connector.generatePodName()
	if err != nil {
		return nil, err
	}

	connector.logger = log.WithFields(log.Fields{
		"connectionType":  "pod",
		"podName":         connector.podName,
		"namespace":       connector.namespace,
		"useNodeAffinity": connector.nodeName != "",
	})

	go waitForSignal(connector)

	return connector, nil
}

func (conn *podSSHConnector) Connect(IPAddress string, port int, username, sshPrivateKey string) error {
	conn.logger = conn.logger.WithFields(log.Fields{
		"ipAddress": IPAddress,
		"username":  username,
	})

	conn.logger.Info("create pod")
	err := conn.createPod(IPAddress, port)
	if err != nil {
		return err
	}
	conn.podCreated = true

	conn.logger.Info("wait for pod to be ready")
	err = conn.waitForPodToBeReady()
	if err != nil {
		return err
	}

	conn.logger.Info("create port forward with kubectl")
	conn.portforwardPID, err = conn.createPortForward()
	if err != nil {
		return err
	}

	for tries := 0; tries < 5; tries++ {
		if conn.shutdown {
			return nil
		}
		conn.logger.Info("try connecting to node through port forward")
		err = conn.invokeSSHThrougPortForward(username, sshPrivateKey)
		if err != nil && tries < 5 {
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (conn *podSSHConnector) Cleanup() {
	conn.shutdown = true

	if !conn.podCreated {
		return
	}

	err := conn.removePod()
	if err != nil {
		conn.logger.Error(err)
	}
}

func (conn *podSSHConnector) Shutdown() {
	conn.shutdown = true
}

func (conn *podSSHConnector) createPortForward() (int, error) {
	commandArgs := []string{
		"--kubeconfig",
		conn.kubeconfig,
		"-n",
		conn.namespace,
		"port-forward",
		conn.podName,
		strconv.Itoa(localPort) + ":2222",
	}

	c := conn.getCommand(kubectlCmd, commandArgs...)
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return 0, err
	}

	return c.Process.Pid, nil
}

func (conn *podSSHConnector) generatePodName() error {
	u, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	conn.podName = fmt.Sprintf("%s-%s", podNamePrefix, u.String())

	return nil
}

func (conn *podSSHConnector) createPod(IPAddress string, port int) error {
	commandArgs := make([]string, 0)
	if conn.nodeName != "" {
		commandArgs = append(commandArgs, []string{
			"--overrides",
			`{"apiVersion":"v1","spec":{"tolerations":[{"effect":"NoExecute","operator":"Exists"},{"effect":"NoSchedule","operator":"Exists"}],"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchFields":[{"key":"metadata.name","operator":"In","values":["` + conn.nodeName + `"]}]}]}}}}}`,
		}...)
	}
	commandArgs = append(commandArgs, []string{
		"--kubeconfig",
		conn.kubeconfig,
		"-n",
		conn.namespace,
		"run",
		"--restart=Never",
		"--generator=run-pod/v1",
		"--grace-period=1",
		"--image=alpine",
		"--command",
		conn.podName,
		"--",
		"sh", "-c", "apk --no-cache add socat && socat TCP-LISTEN:2222,reuseaddr,fork TCP:" + IPAddress + ":" + strconv.Itoa(port),
	}...)

	c := conn.getCommand(kubectlCmd, commandArgs...)
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
	}

	return nil
}

func (conn *podSSHConnector) waitForPodToBeReady() error {
	commandArgs := []string{
		"--kubeconfig",
		conn.kubeconfig,
		"-n",
		conn.namespace,
		"wait",
		"--for=condition=Ready",
		"pod",
		conn.podName,
	}

	c := conn.getCommand(kubectlCmd, commandArgs...)
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
	}

	conn.podRunning = true

	return nil
}

func (conn *podSSHConnector) invokeSSHThrougPortForward(username, sshPrivateKey string) error {
	commandArgs := []string{
		"-4",
		"-o",
		"ConnectTimeout=" + strconv.Itoa(connectionTimeoutInSeconds),
		"-i",
		sshPrivateKey,
		"-l",
		username,
		"-o",
		"StrictHostKeyChecking=no",
		"-o",
		"UserKnownHostsFile=/dev/null",
		"-p",
		strconv.Itoa(localPort),
		"localhost",
	}

	c := conn.getCommand(sshCmd, commandArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		return err
	}

	return nil
}

func (conn *podSSHConnector) removePod() error {
	conn.logger.Info("stop port forwarder")

	proc, err := os.FindProcess(conn.portforwardPID)
	if err != nil {
		return err
	}
	if proc.Pid > 0 {
		err = proc.Signal(os.Interrupt)
		if err != nil {
			return err
		}
	}

	conn.logger.Info("remove pod")

	commandArgs := []string{
		"--kubeconfig",
		conn.kubeconfig,
		"-n",
		conn.namespace,
		"delete",
		"po",
		conn.podName,
	}
	c := conn.getCommand(kubectlCmd, commandArgs...)
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		conn.logger.Error(err)
		return nil
	}

	return nil
}

func (conn *podSSHConnector) getCommand(command string, args ...string) *exec.Cmd {
	conn.logger.WithField("command", command+" "+strings.Join(args, " ")).Debug("execute")
	return exec.Command(command, args...)
}
