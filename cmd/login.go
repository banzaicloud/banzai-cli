// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var options struct {
	endpoint string
	username string
	password string
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure and log in to a Banzai Cloud context",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("login called")
		reader := bufio.NewReader(os.Stdin)

		options.username = os.Getenv("USER")
		fmt.Printf("Please provide your Github credentials to retrieve and store a token for using Pipeline.\n\n")
		fmt.Printf("Github username [%s]: ", options.username)
		if username, err := reader.ReadString('\n'); err != nil {
			log.Fatalf("could not read username: %v", err)
		} else if username != "\n" {
			options.username = strings.TrimSpace(username)
		}

		fmt.Print("Github password: ")
		if password, err := terminal.ReadPassword(int(syscall.Stdin)); err != nil {
			log.Fatalf("could not read password: %v", err)
		} else {
			options.password = string(password)
		}
		fmt.Println()
		fmt.Printf("%+v\n", options)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&options.endpoint, "endpoint", "e", "https://beta.banzaicloud.io/pipeline", "The endpoint of the Banzai Cloud Pipeline instance to use")
	loginCmd.MarkFlagRequired("endpoint")
}
