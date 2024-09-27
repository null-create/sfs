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

# Create bin directory
BUILD_DIR="./bin"
if [ ! -d BUILD_DIR ]; then
  mkdir -p "$BUILD_DIR"
fi

# Create final executable name
if [ "$GOOS" == "windows" ]; then
  OUTPUT_FILE="$PROJECT_NAME.exe"
else
  OUTPUT_FILE="$PROJECT_NAME"
fi

# Add empty .env file to be populated during setup
configs='
ADMIN_MODE=""
BUFFERED_EVENTS=""
CLIENT_ADDRESS=""
CLIENT_BACKUP_DIR=""
CLIENT_EMAIL=""
CLIENT_HOST=""
CLIENT_ID=""
CLIENT_LOCAL_BACKUP=""
CLIENT_LOG_DIR=""
CLIENT_NAME=""
CLIENT_NEW_SERVICE=""
CLIENT_PASSWORD=""
CLIENT_PORT=""
CLIENT_PROFILE_PIC=""
CLIENT_ROOT=""
CLIENT_TESTING=""
CLIENT_USERNAME=""
EVENT_BUFFER_SIZE=""
JWT_SECRET=""
NEW_SERVICE=""
SERVER_ADDR=""
SERVER_ADMIN=""
SERVER_ADMIN_KEY=""
SERVER_HOST=""
SERVER_LOG_DIR=""
SERVER_PORT=""
SERVER_TIMEOUT_IDLE=""
SERVER_TIMEOUT_READ=""
SERVER_TIMEOUT_WRITE=""
SERVICE_ENV=""
SERVICE_LOG_DIR=""
SERVICE_ROOT=""
SERVICE_TEST_ROOT=""
'

if [ ! -f ".env" ]; then
  echo $configs >".env"
fi

# Build the project
echo "Downloading dependencies..."
go mod download
go mod tidy

echo "Building project for $GOOS/$GOARCH..."
GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT_FILE"

cp $OUTPUT_FILE $BUILD_DIR/$OUTPUT_FILE
rm $OUTPUT_FILE

if ! ./bin/"$OUTPUT_FILE" -h; then
  echo "Failed to build new binary for $GOOS/$GOARCH"
  exit 1
fi
