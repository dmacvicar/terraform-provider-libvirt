// Copyright 2016 CoreOS, Inc.
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

package types

import (
	"errors"
	"net/url"

	"github.com/vincent-petithory/dataurl"
)

var (
	ErrInvalidScheme = errors.New("invalid url scheme")
)

func validateURL(s string) error {
	// Empty url is valid, indicates an empty file
	if s == "" {
		return nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "http", "https", "oem", "tftp", "s3":
		return nil
	case "data":
		if _, err := dataurl.DecodeString(s); err != nil {
			return err
		}
		return nil
	default:
		return ErrInvalidScheme
	}
}
