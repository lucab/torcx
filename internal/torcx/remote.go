// Copyright 2018 CoreOS Inc.
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

package torcx

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// EvaluateURL evaluates the URL template for a remote
// and performs variables substitution sourcing values from
// `/etc/os-release`.
func (r *Remote) EvaluateURL() (string, error) {
	if r == nil {
		return "", errors.New("nil Remote")
	}
	if r.TemplateURL == "" {
		return "", errors.New("empty remote URL template")
	}

	vars, any := needSubstitution(r.TemplateURL)
	if !any {
		return r.TemplateURL, nil
	}

	fp, err := os.Open(OsReleasePath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open %s", OsReleasePath)
	}
	defer fp.Close()
	osMeta, err := parseOsRelease(fp)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %s", OsReleasePath)
	}

	// TODO(lucab): this is suboptimal and repetitive
	url := r.TemplateURL
	if vars.board {
		key := "COREOS_BOARD"
		label := fmt.Sprintf("${%s}", key)
		value := osMeta[key]
		if value == "" {
			return "", errors.Errorf("missing required %s value", key)
		}
		url = strings.Replace(url, label, value, -1)

	}
	if vars.vendor {
		key := "ID"
		label := fmt.Sprintf("${%s}", key)
		value := osMeta[key]
		if value == "" {
			return "", errors.Errorf("missing required %s value", key)
		}
		url = strings.Replace(url, label, value, -1)
	}
	if vars.version {
		key := "VERSION_ID"
		label := fmt.Sprintf("${%s}", key)
		value := osMeta[key]
		if value == "" {
			return "", errors.Errorf("missing required %s value", key)
		}
		url = strings.Replace(url, label, value, -1)
	}

	return url, nil
}

// subs describes which variables need to be substituted
// in a URL template.
type subs struct {
	board   bool
	vendor  bool
	version bool
}

// needSubstitution checks whether a URL template contains any
// variables that need to be evaluated.
func needSubstitution(template string) (subs, bool) {
	any := false
	vars := subs{}

	if strings.Contains(template, "${COREOS_BOARD}") {
		vars.board = true
		any = true
	}
	if strings.Contains(template, "${VERSION_ID}") {
		vars.version = true
		any = true
	}
	if strings.Contains(template, "${ID}") {
		vars.vendor = true
		any = true
	}

	return vars, any
}

// parseOsRelease is the parser for os-release.
func parseOsRelease(rd io.Reader) (map[string]string, error) {
	meta := map[string]string{}

	sc := bufio.NewScanner(rd)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == "" {
			continue
		}
		value := strings.Trim(parts[1], `"`)
		if value == "" {
			continue
		}
		meta[parts[0]] = value
	}
	if sc.Err() != nil {
		return meta, sc.Err()
	}
	return meta, nil
}
