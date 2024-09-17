PROJECT_NAME := sfs
BUILD_DIR := ./bin
BIN_NAME := $(BUILD_DIR)/$(PROJECT_NAME)
SRC_DIR := ./cmd/$(PROJECT_NAME)
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Default target
all: build

# Build the project
build: $(GO_FILES) | $(BUILD_DIR)
	@echo "Building project..."
	go build -o $(BIN_NAME) $(SRC_DIR)
	@echo "Build completed: $(BIN_NAME)"

# Compile the project
compile: clean build
	@echo "Compiling project..."
	go build -ldflags="-s -w" -o $(BIN_NAME) $(SRC_DIR)
	@echo "Compilation completed: $(BIN_NAME)"

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Update dependencies
update:
	@echo "Updating dependencies..."
	go mod tidy
	go mod vendor
	@echo "Dependencies updated."

# Clean the build directory
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean completed."

# Create build directory if it doesn't exist
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# Run the project
run: build
	@echo "Running project..."
	$(BIN_NAME)

# Phony targets
.PHONY: all build compile test update clean run
