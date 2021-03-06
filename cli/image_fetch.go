// Copyright 2017 CoreOS Inc.
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

package cli

import (
	"encoding/json"
	"os"

	"github.com/coreos/torcx/pkg/torcx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cmdImageFetch = &cobra.Command{
		Use:   "fetch",
		Short: "Locally fetch a remote torcx image",
		RunE:  runImageFetch,
	}
)

func init() {
	cmdImage.AddCommand(cmdImageFetch)
}

func runImageFetch(cmd *cobra.Command, args []string) error {
	var err error

	if len(args) != 1 {
		return errors.New("missing image/reference")
	}
	refIn := args[0]

	commonCfg, err := fillCommonRuntime()
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	userStorePath := commonCfg.UserStorePath()
	err = os.MkdirAll(userStorePath, 0755)
	if err != nil {
		return err
	}

	storeCache, err := torcx.NewStoreCache(commonCfg.StorePaths)
	if err != nil {
		return err
	}

	archive, err := torcx.DockerFetch(storeCache, userStorePath, refIn)
	if err != nil {
		return err
	}

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")
	err = jsonOut.Encode(archive)

	return err
}
