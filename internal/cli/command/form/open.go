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

package form

import (
	"encoding/json"
	"net"
	"net/http"
	"path"
	"strconv"

	"github.com/banzaicloud/banzai-cli/internal/cli"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

type openOptions struct {
	configFile  string
	openBrowser bool
	port        int
}

// NewOpenCommand creates a new cobra.Command for `banzai form open`.
func NewOpenCommand(banzaiCli cli.Cli) *cobra.Command {
	options := openOptions{}

	cmd := &cobra.Command{
		Use:   "open [config yaml]",
		Short: "Open form",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.configFile = args[0]
			runOpen(banzaiCli, options)
		},
	}

	cmd.Flags().BoolVar(&options.openBrowser, "browser", false, "open browser form")
	cmd.Flags().IntVar(&options.port, "port", 0 /* find an open port */, "port number")

	return cmd
}

func runOpen(banzaiCli cli.Cli, options openOptions) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(options.port))
	if err != nil {
		log.Fatal(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	log.Debugf("using port %d\n", port)

	web := packr.New("web", path.Join(".", "web"))
	http.Handle("/", http.FileServer(web))
	http.HandleFunc("/api/v1/form", createHandler(options))

	if options.openBrowser {
		open.Start("http://127.0.0.1:" + strconv.Itoa(port))
	} else {
		log.Infof("to access the form navigate to http://127.0.0.1:%d using a web browser", port)
	}

	log.Fatal(http.Serve(listener, nil))
}

func createHandler(options openOptions) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			file, err := readConfig(options.configFile)
			if err != nil {
				log.Error(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			b, err := json.Marshal(file.Form)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Write(b)
			return
		}

		if r.Method == "POST" {
			file, err := readConfig(options.configFile)
			if err != nil {
				log.Fatal(err)
			}

			var values map[string]interface{}
			decoder := json.NewDecoder(r.Body)
			err = decoder.Decode(&values)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, group := range file.Form {
				for _, field := range group.Fields {
					field.Value = values[field.Key]
				}
			}

			err = writeConfig(options.configFile, file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
