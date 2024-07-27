#!/usr/bin/env bash

set -eu

download_deps() {
  go mod download
  go mod tidy
}

build() {
  echo "building SFS binary for $1 $2 ..."
  GOOS=$1 GOARCH=$2 go build -o $3
}

# check if go is installed first
if ! command -v go &>/dev/null; then
  echo "golang is not installed"
  echo "install before running the build script"
  exit 1
fi

# download and install dependencies
download_deps

# build executable based on the host OS
case "$(uname -s)" in
Linux*)
  # TODO: refactor this to call each of these depending
  # on architecture
  OUT_FILE="sfs"
  build linux linux-amd64 $OUT_FILE
  ;;
Darwin*)
  OUT_FILE="sfs"
  build darwin mac-amd64 $OUT_FILE
  ;;
CYGWIN* | MINGW32* | MSYS* | MINGW*)
  OUT_FILE="sfs.exe"
  build windows amd64 $OUT_FILE
  ;;
*)
  echo "Unsupported operating system."
  exit 1
  ;;
esac

# set path varible for sfs CLI, then test
# BINPATH="$($PWD)/${OUT_FILE}"
# export PATH="$PATH:${BINPATH}"

./sfs -h
if [[ $? -ne 0 ]]; then
  echo "failed to create executable"
  exit 1
fi

exit 0
