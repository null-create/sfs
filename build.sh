#!/usr/bin/env bash

PROJECT_NAME="sfs"

# Determine architecture
ARCH=$(uname -m)
case "$ARCH" in
x86_64)
  GOARCH="amd64"
  ;;
i386 | i686)
  GOARCH="386"
  ;;
aarch64 | arm64)
  GOARCH="arm64"
  ;;
armv7l)
  GOARCH="arm"
  ;;
ppc64le)
  GOARCH="ppc64le"
  ;;
s390x)
  GOARCH="s390x"
  ;;
*)
  echo "Unsupported architecture: $ARCH"
  exit 1
  ;;
esac

# Determine the OS
OS=$(uname -s)
case "$OS" in
Linux*)
  GOOS="linux"
  ;;
Darwin*)
  GOOS="darwin"
  ;;
CYGWIN* | MINGW32* | MSYS* | MINGW*)
  GOOS="windows"
  ;;
*)
  echo "Unsupported OS: $OS"
  exit 1
  ;;
esac

# Create the output directory
BUILD_DIR="./bin"
if [ -d BUILD_DIR ]; then
  mkdir -p "$BUILD_DIR"
fi

# Output binary name based on OS
if [ "$GOOS" == "windows" ]; then
  OUTPUT_FILE="$PROJECT_NAME.exe"
else
  OUTPUT_FILE="$PROJECT_NAME"
fi

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod tidy

# Build the project
echo "Building project for $GOOS/$GOARCH..."
GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT_FILE"

cp $OUTPUT_FILE $BUILD_DIR/$OUTPUT_FILE
rm $OUTPUT_FILE

if ! ./bin/"$OUTPUT_FILE" -h; then
  echo "Failed to build new binary for $GOOS/$GOARCH"
  exit 1
fi
