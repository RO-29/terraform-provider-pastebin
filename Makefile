.PHONY: build build-terraform clean test testacc

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get
GOTEST = $(GOCMD) test
TERRAFORM_BINARY_NAME = terraform-provider-pastebin


TERRAFORM_PROVIDER_VERSION = 1.0.0

# Build directory
BUILD_DIR = build

# Default target
all: build

# Build Terraform provider
build: build-terraform

# Build the Terraform provider
build-terraform:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(TERRAFORM_BINARY_NAME) $(TERRAFORM_MAIN_PATH)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Download dependencies
deps:
	$(GOGET) -d ./...

# Install Terraform provider locally for development
install-terraform-provider: build-terraform
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/RO-29/pastebin/1.0.0/$(shell go env GOOS)_$(shell go env GOARCH)
	cp $(BUILD_DIR)/$(TERRAFORM_BINARY_NAME) ~/.terraform.d/plugins/registry.terraform.io/RO-29/pastebin/1.0.0/$(shell go env GOOS)_$(shell go env GOARCH)/$(TERRAFORM_BINARY_NAME)_$(TERRAFORM_PROVIDER_VERSION)

# Run unit tests
test:
	$(GOTEST) -v ./...

# Run acceptance tests
testacc:
	TF_ACC=1 $(GOTEST) -v ./... -timeout 120m

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run tests in short mode (excludes long-running tests)
test-short:
	$(GOTEST) -v -short ./...

# Run specific test
test-run:
	@read -p "Enter test name pattern: " test_pattern; \
	$(GOTEST) -v ./... -run $$test_pattern
