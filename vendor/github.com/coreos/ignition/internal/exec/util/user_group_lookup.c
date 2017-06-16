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
#include <grp.h>
#include <sched.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

#include "user_group_lookup.h"

#define STACK_SIZE (64 * 1024)

/* This is all a bit copy-and-pasty from update-ssh-keys/authorized_keys_d,
 * TODO(vc): refactor authorized_keys_d a bit so external packages can reuse
 * the pieces duplicated here.
 */
typedef struct lookup_ctxt {
	void			*stack;

	const char		*name;
	const char		*root;

	lookup_res_t	*res;
	int			ret;
	int			err;
} lookup_ctxt_t;

enum lookup_type {
	LOOKUP_TYPE_USER,
	LOOKUP_TYPE_GROUP
};


static int user_lookup_fn(lookup_ctxt_t *ctxt) {
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

static int group_lookup_fn(lookup_ctxt_t *ctxt) {
	char buf[16 * 1024];
	struct group g, *gptr;

	if(chroot(ctxt->root) == -1) {
		goto out_err;
	}

	if(getgrnam_r(ctxt->name, &g, buf, sizeof(buf), &gptr) != 0 || !gptr) {
		goto out_err;
	}

	if(!(ctxt->res->name = strdup(g.gr_name))) {
		goto out_err;
	}

	ctxt->res->gid = g.gr_gid;

	return 0;

out_err:
	ctxt->err = errno;
	ctxt->ret = -1;
	return 0;
}

int lookup(const char *root, const char *name, lookup_res_t *res, enum lookup_type lt) {
	lookup_ctxt_t	ctxt = {
					.root = root,
					.name = name,
					.res = res,
					.ret = 0
				};
	int			pid, ret = 0;
	sigset_t		allsigs, orig;
	int(*fn)(lookup_ctxt_t *);

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

	switch(lt) {
		case LOOKUP_TYPE_USER:
			fn = user_lookup_fn;
			break;
		case LOOKUP_TYPE_GROUP:
			fn = group_lookup_fn;
			break;
	}

	pid = clone((int(*)(void *))fn, ctxt.stack + STACK_SIZE, CLONE_VM, &ctxt);

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

/* user_lookup() looks up a user in a chroot.
 * returns 0 and the results in res on success,
 * res->name will be NULL if user doesn't exist.
 * returns -1 on error.
 */
int user_lookup(const char *root, const char *name, lookup_res_t *res) {
	return lookup(root, name, res, LOOKUP_TYPE_USER);
}

/* group_lookup() looks up a group in a chroot.
 * returns 0 and the results in res on success,
 * res->name will be NULL if group doesn't exist.
 * returns -1 on error.
 */
int group_lookup(const char *root, const char *name, lookup_res_t *res) {
	return lookup(root, name, res, LOOKUP_TYPE_GROUP);
}

/* user_lookup_res_free() frees any memory allocated by a successful user_lookup(). */
void user_lookup_res_free(lookup_res_t *res) {
	free(res->home);
	free(res->name);
}

/* group_lookup_res_free() frees any memory allocated by a successful group_lookup(). */
void group_lookup_res_free(lookup_res_t *res) {
	free(res->name);
}
