#!/bin/bash -

set -eu

BINPATH="$(pwd)/bin"

build() {
	echo "building SFS binary for $1 $2 ..."
  OUTPUT="${BINPATH}/sfs"
	GOOS=$1 GOARCH=$2 go build -o sfs
}

# build executable based on the hose OS
case "$(uname -s)" in
  Linux*)
    build linux arm linux-arm
		build linux amd64 linux-amd64
		build linux 386 linux-386
    ;;
  Darwin*)
    build darwin amd64 mac-amd64
    ;;
  CYGWIN*|MINGW32*|MSYS*|MINGW*)
    build windows amd64 win-amd64.exe
    ;;
  *)
    echo "Unsupported operating system."
    exit 1
    ;;
esac

cp sfs "${BINPATH}/sfs"
rm sfs

# set path varible for sfs CLI, then test
export PATH="${PATH:+${PATH}:}${BINPATH}"

sfs -h
if [[ $? -ne 0 ]]; then
  echo "failed to set PATH variable"
  exit 1
fi

exit 0