# Platform-agnostic Makefile for gh-action-readme
# Usage: make help

SHELL := /bin/bash

# Detect OS and architecture
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
OS := $(shell echo $(UNAME_S) | tr '[:upper:]' '[:lower:]')
ARCH := $(shell echo $(UNAME_M) | tr '[:upper:]' '[:lower:]')

# Tool versions (can override on command line)
GOLANGCI_LINT_VERSION ?= v1.54.2

# Tool commands (can override on command line)
GOLANGCI_LINT ?= golangci-lint
YAMLLINT ?= yamllint
MARKDOWNLINT ?= markdownlint

.PHONY: help install-tools lint lint-go lint-yaml lint-md test build clean clear check

help:
	@echo "Available Makefile commands:"
	@echo "  make install-tools   Install all required linting and formatting tools"
	@echo "  make lint            Run all linters (Go, YAML, Markdown)"
	@echo "  make lint-go         Run golangci-lint"
	@echo "  make lint-yaml       Run yamllint"
	@echo "  make lint-md         Run markdownlint"
	@echo "  make test            Run all Go tests"
	@echo "  make build           Build the application"
	@echo "  make clean           Remove build/test artifacts"
	@echo "  make clear           Remove temporary files"
	@echo "  make check           Run linting and testing"
	@echo ""
	@echo "Variables:"
	@echo "  GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION)"
	@echo "  GOLANGCI_LINT=$(GOLANGCI_LINT)"
	@echo "  YAMLLINT=$(YAMLLINT)"
	@echo "  MARKDOWNLINT=$(MARKDOWNLINT)"
	@echo "  OS=$(OS)"
	@echo "  ARCH=$(ARCH)"

install-tools:
	@echo "Installing required linting tools for $(OS)/$(ARCH)..."
	@if ! command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	else echo "golangci-lint already installed."; fi
	@if ! command -v $(YAMLLINT) >/dev/null 2>&1; then \
		echo "Installing yamllint..."; \
		if command -v pip3 >/dev/null 2>&1; then pip3 install --user yamllint; \
		elif command -v brew >/dev/null 2>&1; then brew install yamllint; \
		else echo "Please install yamllint manually (pip3 or brew required)."; fi \
	else echo "yamllint already installed."; fi
	@if ! command -v $(MARKDOWNLINT) >/dev/null 2>&1; then \
		echo "Installing markdownlint-cli..."; \
		if command -v npm >/dev/null 2>&1; then npm install -g markdownlint-cli; \
		else echo "Please install markdownlint-cli manually (npm required)."; fi \
	else echo "markdownlint already installed."; fi
	@echo "All tools installed."

lint-go:
	@echo "Running golangci-lint..."
	$(GOLANGCI_LINT) run --fix

lint-yaml:
	@echo "Running yamllint..."
	$(YAMLLINT) . --config-file .yamllint.yml

lint-md:
	@echo "Running markdownlint..."
	$(MARKDOWNLINT) . --config .markdownlint.yml

lint:
	@echo "Running all linters..."
	-$(MAKE) lint-go
	-$(MAKE) lint-yaml
	-$(MAKE) lint-md

test:
	go test ./...

build:
	go build .

clean:
	rm -rf dist/ coverage.out coverage.html coverage.ci.out

clear:
	@echo "Removing temporary files..."
	rm -f coverage.out fail.log

check: lint test
	@echo "Linting and testing completed."
