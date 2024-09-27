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
ENV_FILE=".env"

if [ ! -f "$ENV_FILE" ]; then
  echo "ADMIN_MODE=\"\"\n\
BUFFERED_EVENTS=\"\"\n\
CLIENT_ADDRESS=\"\"\n\
CLIENT_BACKUP_DIR=\"\"\n\
CLIENT_EMAIL=\"\"\n\
CLIENT_HOST=\"\"\n\
CLIENT_ID=\"\"\n\
CLIENT_LOCAL_BACKUP=\"\"\n\
CLIENT_LOG_DIR=\"\"\n\
CLIENT_NAME=\"\"\n\
CLIENT_NEW_SERVICE=\"\"\n\
CLIENT_PASSWORD=\"\"\n\
CLIENT_PORT=9090\n\
CLIENT_PROFILE_PIC=\"\"\n\
CLIENT_ROOT=\"\"\n\
CLIENT_TESTING=\"\"\n\
CLIENT_USERNAME=\"\"\n\
EVENT_BUFFER_SIZE=\"\"\n\
JWT_SECRET=\"\"\n\
NEW_SERVICE=\"\"\n\
SERVER_ADDR=\"\"\n\
SERVER_ADMIN=\"\"\n\
SERVER_ADMIN_KEY=\"\"\n\
SERVER_HOST=\"\"\n\
SERVER_LOG_DIR=\"\"\n\
SERVER_PORT=9191\n\
SERVER_TIMEOUT_IDLE=\"\"\n\
SERVER_TIMEOUT_READ=\"\"\n\
SERVER_TIMEOUT_WRITE=\"\"\n\
SERVICE_ENV=\"C\"\n\
SERVICE_LOG_DIR=\"\"\n\
SERVICE_ROOT=\"\"\n\
SERVICE_TEST_ROOT=\"\"" >"$ENV_FILE"

  echo ".env file has been generated at $PWD/$ENV_FILE"
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
