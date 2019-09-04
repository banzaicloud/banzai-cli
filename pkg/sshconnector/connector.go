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
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	connectionTimeoutInSeconds = 5
)

type SSHConnector interface {
	Connect(IPAddress string, port int, username, sshPrivateKey string) error
	Cleanup()
	Shutdown()
}

func waitForSignal(conn SSHConnector) {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	// Wait interrupt signal
	sig := <-irqSig
	log.WithField("signal", sig).Debug("initiating shutdown")

	conn.Shutdown()
}
