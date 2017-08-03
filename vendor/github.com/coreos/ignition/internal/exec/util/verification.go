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
	"hash"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/internal/resource"
)

func AssertValid(verify types.Verification, data []byte) error {
	if hash := verify.Hash; hash != nil {
		hashFunc, hashSum, err := verify.HashParts()
		if err != nil {
			return err
		}

		var sum []byte
		switch hashFunc {
		case "sha512":
			rawSum := sha512.Sum512(data)
			sum = rawSum[:]
		default:
			return types.ErrHashUnrecognized
		}

		encodedSum := make([]byte, hex.EncodedLen(len(sum)))
		hex.Encode(encodedSum, sum)
		if string(encodedSum) != hashSum {
			return resource.ErrHashMismatch{
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

	function, _, err := verify.HashParts()
	if err != nil {
		return nil, err
	}

	switch function {
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, types.ErrHashUnrecognized
	}
}
