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

#include <blkid/blkid.h>
#include "blkid.h"

result_t blkid_lookup(const char *device, const char *field_name, char buf[], size_t buf_len)
{
	const char *field_val = "\0";

	blkid_probe pr = blkid_new_probe_from_filename(device);
	if (!pr)
		return RESULT_OPEN_FAILED;

	if (blkid_do_probe(pr) != 0)
		return RESULT_PROBE_FAILED;

	if (blkid_probe_has_value(pr, field_name))
		if (blkid_probe_lookup_value(pr, field_name, &field_val, NULL))
			return RESULT_LOOKUP_FAILED;

	strncpy(buf, field_val, buf_len);

	blkid_free_probe(pr);

	return RESULT_OK;
}
