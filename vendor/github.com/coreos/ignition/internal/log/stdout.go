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

package log

import (
	"fmt"
)

type Stdout struct{}

func (Stdout) Emerg(msg string) error   { fmt.Println("EMERGENCY:", msg); return nil }
func (Stdout) Alert(msg string) error   { fmt.Println("ALERT    :", msg); return nil }
func (Stdout) Crit(msg string) error    { fmt.Println("CRITICAL :", msg); return nil }
func (Stdout) Err(msg string) error     { fmt.Println("ERROR    :", msg); return nil }
func (Stdout) Warning(msg string) error { fmt.Println("WARNING  :", msg); return nil }
func (Stdout) Notice(msg string) error  { fmt.Println("NOTICE   :", msg); return nil }
func (Stdout) Info(msg string) error    { fmt.Println("INFO     :", msg); return nil }
func (Stdout) Debug(msg string) error   { fmt.Println("DEBUG    :", msg); return nil }
func (Stdout) Close() error             { return nil }
