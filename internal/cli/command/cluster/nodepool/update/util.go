// Copyright © 2020 Banzai Cloud
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

package update

import (
	"context"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/banzai-cli/internal/cli"
)

func checkUpdateProcess(banzaiCli cli.Cli, processID string) error {
	p, resp, err := banzaiCli.Client().ProcessesApi.GetProcess(context.Background(), banzaiCli.Context().OrganizationID(), processID)
	if err != nil {
		return errors.WrapIf(err, "failed to read node pool update process")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err := errors.NewWithDetails("reading node pool update process failed with http status code", "status_code", resp.StatusCode)
		return err
	}

	if !strings.HasSuffix(p.Type, "update-node-pool") {
		return errors.New("not a node pool update process")
	}

	return nil
}
