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
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type directSSHConnector struct {
	logger log.FieldLogger
}

func NewDirectSSHConnector() *directSSHConnector {
	connector := &directSSHConnector{
		logger: log.WithField("connectionType", "direct"),
	}

	go waitForSignal(connector)

	return connector
}

func (conn *directSSHConnector) Connect(IPAddress string, port int, username, sshPrivateKey string) error {
	conn.logger = conn.logger.WithFields(log.Fields{
		"ipAddress": IPAddress,
		"username":  username,
	})

	commandArgs := []string{
		"-o",
		"ConnectTimeout=" + strconv.Itoa(connectionTimeoutInSeconds),
		"-o",
		"StrictHostKeyChecking=no",
		"-o",
		"UserKnownHostsFile=/dev/null",
		"-i",
		sshPrivateKey,
		"-l",
		username,
		"-p",
		strconv.Itoa(port),
		IPAddress,
	}

	conn.logger.Info("connecting to node")

	c := conn.getCommand("ssh", commandArgs...)
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
	}

	return nil
}

func (conn *directSSHConnector) Cleanup()  {}
func (conn *directSSHConnector) Shutdown() {}

func (conn *directSSHConnector) getCommand(command string, args ...string) *exec.Cmd {
	conn.logger.WithField("command", command+" "+strings.Join(args, " ")).Debug("execute")
	return exec.Command(command, args...)
}
