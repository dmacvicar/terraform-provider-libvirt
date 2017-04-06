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
	"encoding/json"
	"fmt"
	"regexp"
)

type Partition struct {
	Label    PartitionLabel     `json:"label,omitempty"`
	Number   int                `json:"number"`
	Size     PartitionDimension `json:"size"`
	Start    PartitionDimension `json:"start"`
	TypeGUID PartitionTypeGUID  `json:"typeGuid,omitempty"`
}

type PartitionLabel string
type partitionLabel PartitionLabel

func (n *PartitionLabel) UnmarshalJSON(data []byte) error {
	tn := partitionLabel(*n)
	if err := json.Unmarshal(data, &tn); err != nil {
		return err
	}
	*n = PartitionLabel(tn)
	return n.AssertValid()
}

func (n PartitionLabel) AssertValid() error {
	// http://en.wikipedia.org/wiki/GUID_Partition_Table#Partition_entries:
	// 56 (0x38) 	72 bytes 	Partition name (36 UTF-16LE code units)

	// XXX(vc): note GPT calls it a name, we're using label for consistency
	// with udev naming /dev/disk/by-partlabel/*.
	if len(string(n)) > 36 {
		return fmt.Errorf("partition labels may not exceed 36 characters")
	}
	return nil
}

type PartitionDimension uint64

func (n *PartitionDimension) UnmarshalJSON(data []byte) error {
	var pd uint64
	if err := json.Unmarshal(data, &pd); err != nil {
		return err
	}
	*n = PartitionDimension(pd)
	return nil
}

type PartitionTypeGUID string
type partitionTypeGUID PartitionTypeGUID

func (d *PartitionTypeGUID) UnmarshalJSON(data []byte) error {
	td := partitionTypeGUID(*d)
	if err := json.Unmarshal(data, &td); err != nil {
		return err
	}
	*d = PartitionTypeGUID(td)
	return d.AssertValid()
}

func (d PartitionTypeGUID) AssertValid() error {
	ok, err := regexp.MatchString("^(|[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12})$", string(d))
	if err != nil {
		return fmt.Errorf("error matching type-guid regexp: %v", err)
	}
	if !ok {
		return fmt.Errorf(`partition type-guid must have the form "01234567-89AB-CDEF-EDCB-A98765432101", got: %q`, string(d))
	}
	return nil
}
