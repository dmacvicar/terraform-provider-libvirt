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

#define _GNU_SOURCE
#include <errno.h>
#include <fcntl.h>
#include <sched.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

#include "as_user.h"

// By doing these operations in CGO, we prevent the Go scheduler from
// potentially executing while our uid/gid/umask is changed.
//
// By doing them in a new !CLONE_FS thread we also localize the effects of the
// umask() switching.
//
// The calling go code must supply the appropriate args struct.
// We create a new thread passing the args down verbatim, it's all fairly
// mechanical.  There's no need to restore umask/eids since we just exit the
// created thread when done.
//
// The return value and errno is simply fed back to the caller from the thread
// through some variables in ctxt.  Synchronization is achieved via waitpid().

#define STACK_SIZE (64 * 1024)

typedef struct au_thread_ctxt {
	au_ids_t	*ids;

	void		*stack;
	int		(*fn)(void *);
	void		*fn_args;

	int		ret;
	int		err;
} au_thread_ctxt_t;

/* set_eids() sets the effective gid and uid of the calling thread to ids */
static int set_eids(au_ids_t *ids) {
	uid_t	cu;
	gid_t	cg;

	umask(077);

	cu = geteuid();
	cg = getegid();

	if(cg != ids->gid && setregid(-1, ids->gid) == -1)
		return -1;

	if(cu != ids->uid && setreuid(-1, ids->uid) == -1)
		return -1;

	return 0;
}

/* in_thread_fn() switches users and calls ctxt->fn */
static int in_thread_fn(au_thread_ctxt_t *ctxt) {
	if((ctxt->ret = set_eids(ctxt->ids)) == 0)
		ctxt->ret = ctxt->fn(ctxt->fn_args);

	ctxt->err = errno;

	return 0;
}

/* in_thread() calls in_thread_fn() in a new thread passing fn and args along in
 * via an au_thread_ctxt_t.
 */
static int in_thread(au_ids_t *ids, int (*fn)(void *), void *fn_args) {
	int			pid, ret = 0;
	au_thread_ctxt_t	ctxt = {
					.ids = ids,
					.fn = fn,
					.fn_args = fn_args,
					.err = 0,
					.ret = -1,
				};
	sigset_t		allsigs, orig;

	if(!(ctxt.stack = malloc(STACK_SIZE))) {
		ret = -1;
		goto out;
	}

	/* It's necessary to block all signals before cloning, so the child
	 * doesn't run any of the Go runtime's signal handlers.
	 */
	if((ret = sigemptyset(&orig)) == -1 ||
	   (ret = sigfillset(&allsigs)) == -1)
		goto out_stack;

	if((ret = sigprocmask(SIG_BLOCK, &allsigs, &orig)) == -1)
		goto out_stack;

	pid = clone((int(*)(void *))in_thread_fn, ctxt.stack + STACK_SIZE,
		    CLONE_FILES|CLONE_VM, &ctxt);

	ret = sigprocmask(SIG_SETMASK, &orig, NULL);

	if(pid != -1) {
		if(waitpid(pid, NULL, __WCLONE) == -1 && errno != ECHILD) {
			ret = -1;
			goto out_stack;
		}
	} else {
		ret = -1;
	}

	if(ret != -1) {
		errno = ctxt.err;
		ret = ctxt.ret;
	}

out_stack:
	free(ctxt.stack);

out:
	return ret;
}

/* au_open() */
static int au_open_fn(au_open_args_t *args) {
	return open(args->path, args->flags, args->mode);
}

int au_open(au_open_args_t *args) {
	return in_thread(&args->ids, (int(*)(void *))au_open_fn, args);
}

/* au_mkdir_all() */
static int au_mkdir_all_fn(au_mkdir_all_args_t *args) {
	int	ret = 0;
	char	*sep, *dup;

	/* TODO(vc): rewrite this, it's difficult to follow and probably buggy */

	if(!*(args->path))
		goto out;

	if(!(dup = strdup(args->path))) {
		ret = -1;
		goto out;
	}

	for(sep = dup + 1; sep;) {
		/* find next slash or end of path */
		while(*sep && (*sep) != '/') sep++;
		if(*sep)
			*sep = '\0';
		else
			sep = NULL;

		if(((ret = mkdir(dup, args->mode)) == -1) && errno != EEXIST)
			goto out_free;

		if(sep) {
			/* restore the '/' and skip '/'s */
			*sep = '/';
			while(*sep && *sep == '/') sep++;
		}
	}

	if(ret == -1 && errno == EEXIST)
		ret = 0;

out_free:
	free(dup);

out:
	return ret;
}

int au_mkdir_all(au_mkdir_all_args_t *args) {
	return in_thread(&args->ids, (int(*)(void *))au_mkdir_all_fn, args);
}

/* au_rename() */
static int au_rename_fn(au_rename_args_t *args) {
	return rename(args->oldpath, args->newpath);
}

int au_rename(au_rename_args_t *args) {
	return in_thread(&args->ids, (int(*)(void *))au_rename_fn, args);
}
