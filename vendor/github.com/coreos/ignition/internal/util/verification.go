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

package util

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"

	"github.com/coreos/ignition/internal/config/types"
)

var (
	ErrHashMalformed    = errors.New("malformed hash specifier")
	ErrHashUnrecognized = errors.New("unrecognized hash function")
)

// ErrHashMismatch is returned when the calculated hash for a fetched object
// doesn't match the expected sum of the object.
type ErrHashMismatch struct {
	Calculated string
	Expected   string
}

func (e ErrHashMismatch) Error() string {
	return fmt.Sprintf("hash verification failed (calculated %s but expected %s)",
		e.Calculated, e.Expected)
}

// HashParts will return the sum and function (in that order) of the hash stored
// in this Verification, or an error if there is an issue during parsing.
func HashParts(v types.Verification) (string, string, error) {
	if v.Hash == nil {
		// The hash can be nil
		return "", "", nil
	}
	parts := strings.SplitN(*v.Hash, "-", 2)
	if len(parts) != 2 {
		return "", "", ErrHashMalformed
	}

	return parts[0], parts[1], nil
}

func AssertValid(verify types.Verification, data []byte) error {
	if hash := verify.Hash; hash != nil {
		hashFunc, hashSum, err := HashParts(verify)
		if err != nil {
			return err
		}

		var sum []byte
		switch hashFunc {
		case "sha512":
			rawSum := sha512.Sum512(data)
			sum = rawSum[:]
		default:
			return ErrHashUnrecognized
		}

		encodedSum := make([]byte, hex.EncodedLen(len(sum)))
		hex.Encode(encodedSum, sum)
		if string(encodedSum) != hashSum {
			return ErrHashMismatch{
				Calculated: string(encodedSum),
				Expected:   hashSum,
			}
		}
	}

	return nil
}

func GetHasher(verify types.Verification) (hash.Hash, error) {
	if verify.Hash == nil {
		return nil, nil
	}

	function, _, err := HashParts(verify)
	if err != nil {
		return nil, err
	}

	switch function {
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, ErrHashUnrecognized
	}
}
