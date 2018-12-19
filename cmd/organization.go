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
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

const orgIdKey = "organization.id"

// organizationCmd represents the organization command
var organizationCmd = &cobra.Command{
	Use:     "organization",
	Aliases: []string{"org", "orgs"},
	Short:   "List and select organizations.",
	Run:     OrganizationList,
}

var organizationSelectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select organization.",
	Run:   OrganizationSelect,
	Args:  cobra.MaximumNArgs(1),
}

var organizationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations.",
	Run:   OrganizationList,
	Args:  cobra.NoArgs,
}

func OrganizationList(cmd *cobra.Command, args []string) {
	pipeline := InitPipeline()
	orgs, _, err := pipeline.OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		log.Fatalf("could not list organizations: %v", err)
	}
	Out(orgs, []string{"Id", "Name"})
	id := GetOrgId(false)
	for _, org := range orgs {
		if org.Id == id {
			log.Infof("Organization %q (%d) is selected.", org.Name, org.Id)
		}
	}
}

func searchOrganizationId(name string) int32 {
	pipeline := InitPipeline()
	orgs, _, err := pipeline.OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		logAPIError("list organizations", err, nil)
		return 0
	}
	for _, org := range orgs {
		if org.Name == name {
			return org.Id
		}
	}
	log.Errorf("Could not find organization %q", name)
	return 0
}

func searchOrganizationName(id int32) string {
	pipeline := InitPipeline()
	org, _, err := pipeline.OrganizationsApi.GetOrg(context.Background(), id)
	if err != nil {
		logAPIError("get organization", err, id)
		return ""
	}
	return org.Name
}

func OrganizationSelect(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		id := searchOrganizationId(args[0])
		if id > 0 {
			viper.Set(orgIdKey, id)
		}
	} else {
		viper.Set(orgIdKey, 0)
		GetOrgId(true)
	}
	WriteConfig()
}

func GetOrgId(ask bool) int32 {
	id := viper.GetInt32(orgIdKey)
	if id != 0 {
		return id
	}
	if ask && !isInteractive() {
		log.Fatal("No organization is selected. Use the --organization switch, or set it using `banzai org`.")
	}
	pipeline := InitPipeline()
	orgs, _, err := pipeline.OrganizationsApi.ListOrgs(context.Background())
	if err != nil {
		log.Fatalf("could not list organizations: %v", err)
	}
	orgSlice := make([]string, len(orgs))
	for i, org := range orgs {
		orgSlice[i] = org.Name
	}
	name := ""
	survey.AskOne(&survey.Select{Message: "Organization:", Options: orgSlice}, &name, survey.Required)
	id = searchOrganizationId(name)
	if id != 0 {
		viper.Set(orgIdKey, id)
	}
	return id
}

func init() {
	rootCmd.AddCommand(organizationCmd)
	organizationCmd.AddCommand(organizationListCmd)
	organizationCmd.AddCommand(organizationSelectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// organizationCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
