// Copyright 2015 CoreOS, Inc.
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

package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/ignition/config/v1/types"

	"go4.org/errorutil"
)

var (
	ErrVersion     = errors.New("incorrect config version")
	ErrCloudConfig = errors.New("not a config (found coreos-cloudconfig)")
	ErrEmpty       = errors.New("not a config (empty)")
	ErrScript      = errors.New("not a config (found coreos-cloudinit script)")
)

func Parse(rawConfig []byte) (config types.Config, err error) {
	if err = json.Unmarshal(rawConfig, &config); err == nil {
		if config.Version != types.Version {
			err = ErrVersion
		}
	} else if isEmpty(rawConfig) {
		err = ErrEmpty
	} else if isCloudConfig(rawConfig) {
		err = ErrCloudConfig
	} else if isScript(rawConfig) {
		err = ErrScript
	}
	if serr, ok := err.(*json.SyntaxError); ok {
		line, col, highlight := errorutil.HighlightBytePosition(bytes.NewReader(rawConfig), serr.Offset)
		err = fmt.Errorf("error at line %d, column %d\n%s%v", line, col, highlight, err)
	}

	return
}

func isEmpty(userdata []byte) bool {
	return len(userdata) == 0
}
