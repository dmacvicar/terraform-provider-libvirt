#!/usr/bin/env bash

set -e

[ $# == 2 ] || { echo "usage: $0 version commit" && exit 1; }

VER=$1
COMMIT=$2

[[ "${VER}" =~ ^v[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+$ ]] || {
	echo "malformed version: \"${VER}\""
	exit 2
}

[[ "${COMMIT}" =~ ^[[:xdigit:]]+$ ]] || {
	echo "malformed commit id: \"${COMMIT}\""
	exit 3
}

source ./build
go run internal/util/tools/prerelease_check.go

# TODO(vc): generate NEWS as part of the release process.
# @marineam suggested using git notes to associate NEWS-destined payloads
# with objects, we just need to define a syntax and employ them.
# I would like to be able to write the NEWS annotation as part of the commit message,
# while still having it go into a note.

git tag --sign --message "Ignition ${VER}" "${VER}" "${COMMIT}"

git verify-tag --verbose "${VER}"
