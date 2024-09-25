#!/bin/bash

env_file=".env"
prog_src="pkg"
directories=(
  "./$prog_src/auth"
  "./$prog_src/client"
  "./$prog_src/configs"
  "./$prog_src/db"
  "./$prog_src/env"
  "./$prog_src/logger"
  "./$prog_src/monitor"
  "./$prog_src/server"
  "./$prog_src/service"
  "./$prog_src/transfer"
)

# TODO: add defaults for new .env files

if [ ! -f ".env" ]; then
  cat <<EOL >$env_file
ADMIN_MODE="false"
BUFFERED_EVENTS="true"
CLIENT_ADDRESS="localhost:9090"
CLIENT_BACKUP_DIR="C:\\Users\\Jay\\Desktop\\backups"
CLIENT_EMAIL="wnATwienermann.com"
CLIENT_HOST="localhost"
CLIENT_ID="48d0569b-7ac7-11ef-8c04-d85ed3e090f7"
CLIENT_LOCAL_BACKUP="false"
CLIENT_LOG_DIR="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\logger\\logs"
CLIENT_NAME="wienermann nugget"
CLIENT_NEW_SERVICE="false"
CLIENT_PASSWORD="skibityboo"
CLIENT_PORT=9090
CLIENT_PROFILE_PIC="default_profile_pic.jpg"
CLIENT_ROOT="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\client\\run"
CLIENT_TESTING="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\client\\testing"
CLIENT_USERNAME="wiener_nugs_69"
EVENT_BUFFER_SIZE=2
JWT_SECRET="secret-key-goes-here"
NEW_SERVICE="false"
SERVER_ADDR="localhost:9191"
SERVER_ADMIN="admin"
SERVER_ADMIN_KEY="derp"
SERVER_HOST="localhost"
SERVER_LOG_DIR="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\logger\\logs"
SERVER_PORT=9191
SERVER_TIMEOUT_IDLE="900s"
SERVER_TIMEOUT_READ="5s"
SERVER_TIMEOUT_WRITE="10s"
SERVICE_ENV="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\env\\.env"
SERVICE_LOG_DIR="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\logger\\logs"
SERVICE_ROOT="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\server\\run\\sfs_test"
SERVICE_TEST_ROOT="C:\\Users\\Jay\\coding-projects\\sfs\\pkg\\server\\testing"
EOL

  echo ".env file has been generated at $PWD/$env_file"
fi

# copy the .env file to each specified directory
for dir in "${directories[@]}"; do
  if [ -d "$dir" ]; then
    cp "$env_file" "$dir"
  fi
done
