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
#include <pwd.h>
#include <sched.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

#include "user_lookup.h"

#define STACK_SIZE (64 * 1024)

/* This is all a bit copy-and-pasty from update-ssh-keys/authorized_keys_d,
 * TODO(vc): refactor authorized_keys_d a bit so external packages can reuse
 * the pieces duplicated here.
 */
typedef struct user_lookup_ctxt {
	void			*stack;

	const char		*name;
	const char		*root;

	user_lookup_res_t	*res;
	int			ret;
	int			err;
} user_lookup_ctxt_t;


static int user_lookup_fn(user_lookup_ctxt_t *ctxt) {
	char		buf[16 * 1024];
	struct passwd	p, *pptr;

	if(chroot(ctxt->root) == -1) {
		goto out_err;
	}

	if(getpwnam_r(ctxt->name, &p, buf, sizeof(buf), &pptr) != 0 || !pptr) {
		goto out_err;
	}

	if(!(ctxt->res->name = strdup(p.pw_name))) {
		goto out_err;
	}

	if(!(ctxt->res->home = strdup(p.pw_dir))) {
		free(ctxt->res->name);
		goto out_err;
	}

	ctxt->res->uid = p.pw_uid;
	ctxt->res->gid = p.pw_gid;

	return 0;

out_err:
	ctxt->err = errno;
	ctxt->ret = -1;
	return 0;
}

/* user_lookup() looks up a user in a chroot.
 * returns 0 and the results in res on success,
 * res->name will be NULL if user doesn't exist.
 * returns -1 on error.
 */
int user_lookup(const char *root, const char *name, user_lookup_res_t *res) {
	user_lookup_ctxt_t	ctxt = {
					.root = root,
					.name = name,
					.res = res,
					.ret = 0
				};
	int			pid, ret = 0;
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

	pid = clone((int(*)(void *))user_lookup_fn, ctxt.stack + STACK_SIZE,
		    CLONE_VM, &ctxt);

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

/* user_lookup_res_free() frees any memory allocated by a successful user_lookup(). */
void user_lookup_res_free(user_lookup_res_t *res) {
	free(res->home);
	free(res->name);
}
