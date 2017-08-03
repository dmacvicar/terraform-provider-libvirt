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

// The vmware provider fetches a configuration from the VMware Guest Info
// interface.

package vmware

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
)

type config struct {
	data     string
	encoding string
}

func decodeConfig(config config) ([]byte, error) {
	switch config.encoding {
	case "":
		return []byte(config.data), nil

	case "b64", "base64":
		return decodeBase64Data(config.data)

	case "gz", "gzip":
		return decodeGzipData(config.data)

	case "gz+base64", "gzip+base64", "gz+b64", "gzip+b64":
		gz, err := decodeBase64Data(config.data)

		if err != nil {
			return nil, err
		}

		return decodeGzipData(string(gz))
	}

	return nil, fmt.Errorf("unsupported encoding %q", config.encoding)
}

func decodeBase64Data(data string) ([]byte, error) {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode base64: %q", err)
	}

	return decodedData, nil
}

func decodeGzipData(data string) ([]byte, error) {
	reader, err := gzip.NewReader(strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}
