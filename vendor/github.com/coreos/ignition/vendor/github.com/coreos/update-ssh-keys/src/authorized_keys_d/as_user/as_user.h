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

typedef struct au_ids {
	int	uid;
	int	gid;
} au_ids_t;

typedef struct au_open_args {
	au_ids_t	ids;
	const char	*path;
	unsigned int	flags;
	unsigned int	mode;
} au_open_args_t;

typedef struct au_mkdir_all_args {
	au_ids_t	ids;
	const char	*path;
	unsigned int	mode;
} au_mkdir_all_args_t;

typedef struct au_rename_args {
	au_ids_t	ids;
	const char	*oldpath;
	const char	*newpath;
} au_rename_args_t;

int au_open(au_open_args_t *);
int au_mkdir_all(au_mkdir_all_args_t *);
int au_rename(au_rename_args_t *);
