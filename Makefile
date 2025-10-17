# md2pdf Makefile

# Binary name and paths
BINARY_NAME=md2pdf
BIN_DIR=bin
INSTALL_DIR=$(HOME)/.local/bin
CMD_PATH=./cmd/md2pdf

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
VERSION=$(shell date +v%y.%m%d.%H%M)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
BUILD_FLAGS=-v

# Colors for help
BOLD=\033[1m
RESET=\033[0m
CYAN=\033[36m
GREEN=\033[32m

.PHONY: help build install test e2e clean fmt lint tidy run all generate-emoji

# Default target - show help
help: ## Show this help message
	@echo "$(BOLD)md2pdf - Markdown to PDF converter$(RESET)"
	@echo ""
	@echo "$(BOLD)Usage:$(RESET)"
	@echo "  make $(CYAN)<target>$(RESET)"
	@echo ""
	@echo "$(BOLD)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2}'

build: ## Build the binary to bin/
	@echo "$(GREEN)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BIN_DIR)
	@rm -f ./$(BINARY_NAME)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(GREEN)✓ Binary built: $(BIN_DIR)/$(BINARY_NAME)$(RESET)"

install: build ## Install binary to ~/.local/bin (no sudo required)
	@echo "$(GREEN)Installing $(BINARY_NAME) to $(INSTALL_DIR)...$(RESET)"
	@mkdir -p $(INSTALL_DIR)
	@cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installed to $(INSTALL_DIR)/$(BINARY_NAME)$(RESET)"
	@echo ""
	@echo "$(BOLD)Make sure $(INSTALL_DIR) is in your PATH:$(RESET)"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""

test: ## Run unit tests
	@echo "$(GREEN)Running unit tests...$(RESET)"
	$(GOTEST) -v -short ./...

e2e: build ## Run end-to-end tests
	@echo "$(GREEN)Running E2E tests...$(RESET)"
	$(GOTEST) -v -timeout 2m -run TestE2E ./...

clean: ## Remove build artifacts
	@echo "$(GREEN)Cleaning...$(RESET)"
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f *.pdf tests/e2e_*.pdf example-output.pdf
	@echo "$(GREEN)✓ Clean complete$(RESET)"

fmt: ## Format Go code
	@echo "$(GREEN)Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

lint: ## Run linter (requires golangci-lint)
	@echo "$(GREEN)Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/"; \
	fi

tidy: ## Tidy and verify dependencies
	@echo "$(GREEN)Tidying dependencies...$(RESET)"
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "$(GREEN)✓ Dependencies tidied$(RESET)"

run: build ## Build and run with example
	@echo "$(GREEN)Running example...$(RESET)"
	@timeout 20s ./$(BIN_DIR)/$(BINARY_NAME) -i tests/md2pdf_test.md -o example-output.pdf || \
		(echo "$(GREEN)Build timed out or failed$(RESET)" && exit 1)
	@echo "$(GREEN)✓ Generated example-output.pdf$(RESET)"

all: clean tidy fmt build test ## Run clean, tidy, fmt, build and test
	@echo "$(GREEN)✓ All tasks complete$(RESET)"

generate-emoji: ## Generate grayscale emoji PNGs from Twemoji SVGs
	@echo "$(GREEN)Generating emoji assets from Twemoji...$(RESET)"
	@if [ ! -d "../twemoji/assets/svg" ]; then \
		echo "$(BOLD)ERROR:$(RESET) ../twemoji not found"; \
		echo "Clone twemoji repository:"; \
		echo "  cd /Users/aleksander/Documents/projects4"; \
		echo "  git clone https://github.com/jdecked/twemoji.git"; \
		exit 1; \
	fi
	$(GOCMD) run scripts/generate_emoji.go
	@echo "$(GREEN)✓ Emoji assets generated$(RESET)"
