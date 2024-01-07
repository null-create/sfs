#!/bin/bash -

set -eu

VERSION=$(git describe --abbrev=0 --tags)
REVCNT=$(git rev-list --count HEAD)
DEVCNT=$(git rev-list --count $VERSION)
if test $REVCNT != $DEVCNT
then
	VERSION="$VERSION.dev$(expr $REVCNT - $DEVCNT)"
fi
echo "VER: $VERSION"

GITCOMMIT=$(git rev-parse HEAD)
BUILDTIME=$(date -u +%Y/%m/%d-%H:%M:%S)

LDFLAGS="-X main.VERSION=$VERSION -X main.BUILDTIME=$BUILDTIME -X main.GITCOMMIT=$GITCOMMIT"
if [[ -n "${EX_LDFLAGS:-""}" ]]
then
	LDFLAGS="$LDFLAGS $EX_LDFLAGS"
fi

build() {
	echo "building SFS binary for $1 $2 ..."
	GOOS=$1 GOARCH=$2 go build \
		-ldflags "$LDFLAGS" \
		-o bin/sfs-${3:-""}
}

# ----------------------------------------------------------------

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

BINPATH="$(pwd)/bin"

# set path varible for sfs CLI, then test
export PATH=$PATH:$BINPATH

sfs --help

exit 0