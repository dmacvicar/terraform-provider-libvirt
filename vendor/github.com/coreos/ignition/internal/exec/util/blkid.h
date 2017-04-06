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

#ifndef _BLKID_H_
#define _BLKID_H_

#include <string.h>

typedef enum {
	RESULT_OK,
	RESULT_OPEN_FAILED,
	RESULT_PROBE_FAILED,
	RESULT_LOOKUP_FAILED,
} result_t;

result_t blkid_lookup(const char *device, const char *field_name, char buf[], size_t buf_len);

#endif // _BLKID_H_

