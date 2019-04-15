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
	"fmt"
	"os"
	"path"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/banzaicloud/banzai-cli/internal/cli/command"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootOptions struct {
	CfgFile string
	Output  string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "banzai",
	Short:            "A command line client for the Banzai Pipeline platform.",
	PersistentPreRun: preRun,
}

func preRun(cmd *cobra.Command, args []string) {
	if viper.GetBool("output.verbose") {
		log.SetLevel(log.DebugLevel)
	}
}

// Init is a temporary function to set initial values in the root cmd.
func Init(version string, commitHash string, buildDate string, pipelineVersion string) {
	rootCmd.Version = version

	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"Banzai CLI version %s (%s) built on %s\n\nPipeline version: %s\n",
		version,
		commitHash,
		buildDate,
		pipelineVersion,
	))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	flags := rootCmd.PersistentFlags()

	flags.StringVar(&rootOptions.CfgFile, "config", "", "config file (default is $HOME/.banzai/config.yaml)")
	//flags.StringVarP(&BanzaiContext, "context", "c", "default", "name of Banzai Cloud context to use")
	flags.StringVarP(&rootOptions.Output, "output", "o", "default", "output format (default|yaml|json)")

	flags.Int32("organization", 0, "organization id")
	_ = viper.BindPFlag("organization.id", flags.Lookup("organization"))

	flags.Bool("no-color", false, "never display color output")
	_ = viper.BindPFlag("formatting.no-color", flags.Lookup("no-color"))
	flags.Bool("color", false, "use colors on non-tty outputs")
	_ = viper.BindPFlag("formatting.force-color", flags.Lookup("color"))
	flags.Bool("no-interactive", false, "never ask questions interactively")
	_ = viper.BindPFlag("formatting.no-interactive", flags.Lookup("no-interactive"))
	flags.Bool("interactive", false, "ask questions interactively even if stdin or stdout is non-tty")
	_ = viper.BindPFlag("formatting.force-interactive", flags.Lookup("interactive"))

	flags.Bool("verbose", false, "more verbose output")
	_ = viper.BindPFlag("output.verbose", flags.Lookup("verbose"))

	viper.SetDefault("pipeline.basepath", "https://beta.banzaicloud.io/pipeline")

	cli := cli.NewCli(os.Stdout)

	command.AddCommands(rootCmd, cli)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if rootOptions.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(rootOptions.CfgFile)
	} else if envCfg := os.Getenv("BANZAICONFIG"); envCfg != "" {
		// Use config file from BANZAICONFIG env var.
		viper.SetConfigFile(envCfg)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("can't compose the default config file path: %v", err)
		}

		viper.AddConfigPath(path.Join(home, ".banzai"))
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	}
}
