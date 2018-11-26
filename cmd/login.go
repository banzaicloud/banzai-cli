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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var options struct {
	Endpoint string
	Username string
	Password string
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure and log in to a Banzai Cloud context",
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal("Not implemented. Please either set a pipeline token aquired from https://beta.banzaicloud.io/pipeline/api/v1/token in the environemnt variable PIPELINE_TOKEN or as pipeline.token in ~/.banzai/config.yaml")
	}}

/*Run: func(cmd *cobra.Command, args []string) {
	fmt.Println("login called")

	qs := []*survey.Question{}

	if options.Username == "" {
		qs = append(qs, &survey.Question{
			Name: "username",
			Prompt: &survey.Input{
				Message: "Github username:",
				Help:    "Please provide your Github credentials to retrieve and store a token for using Pipeline.",
			},
			Validate:  survey.Required,
			Transform: survey.Title,
		})
	}

	if options.Password == "" {
		qs = append(qs, &survey.Question{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Github password:",
				Help:    "Please provide your Github credentials to retrieve and store a token for using Pipeline. We won't store your password."},
			Validate: survey.Required,
		})
	}

	if len(qs) > 0 {
		if err := survey.Ask(qs, &options); err != nil {
			log.Fatalf("failed to ask for options: %v", err)
		}
	}

	fmt.Printf("%+v\n", options)
},*/

func init() {
	rootCmd.AddCommand(loginCmd)

	/*
		loginCmd.Flags().StringVarP(&options.Endpoint, "endpoint", "e", "https://beta.banzaicloud.io/pipeline", "The endpoint of the Banzai Cloud Pipeline instance to use")
		loginCmd.Flags().StringVarP(&options.Username, "username", "u", "", "Github username to use")
		loginCmd.Flags().StringVarP(&options.Username, "password", "p", "", "Github password to use (not recommended)")
	*/
}
