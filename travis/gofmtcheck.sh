#!/usr/bin/env bash

set -e
GOPKG="libvirt"

lint_pkg () {
  cd $1
  echo "*** checking pkg $1 with gofmt"
  if [ -n "$(go fmt ./...)" ]; then
    echo "Go code on pkg $1 is   not formatted well, run 'go fmt on pkg $1'" 
    exit 1
  fi
  echo " pkg $1 has no lint gofmt errors!"
  cd ..
}

lint_main () {
  echo '*** running gofmt on main.go'
  if [ -n "$(go fmt main.go)" ]; then
    echo "Go code on main.go is not formatted,  please run 'go fmt main.go'" 
    exit 1
  fi
}

echo "==> Checking that code complies with gofmt requirements..."
lint_pkg $GOPKG
echo 
lint_main
echo '==> go fmt OK !!! ***'
exit 0
