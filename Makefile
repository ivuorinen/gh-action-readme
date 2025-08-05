.PHONY: help test lint run example clean readme config-verify security vulncheck audit snyk trivy gitleaks \
	editorconfig editorconfig-fix format

all: help

help: ## Show this help message
	@echo "GitHub Action README Generator - Available Make Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Common workflows:"
	@echo "  make test lint     # Run tests and linting"
	@echo "  make format        # Format code and fix EditorConfig issues"
	@echo "  make security      # Run all security scans"

test: ## Run all tests
	go test ./...

lint: format ## Run linter (after formatting)
	golangci-lint run || true

config-verify: ## Verify golangci-lint configuration
	golangci-lint config verify --verbose

run: ## Run the application
	go run .

example: ## Generate example README
	go run . gen --config config.yaml --output-format=md

readme: ## Generate project README
	go run . gen --config config.yaml --output-format=md

clean: ## Clean build artifacts
	rm -rf dist/

# Code formatting and EditorConfig targets
format: editorconfig-fix ## Format code and fix EditorConfig issues
	@echo "Running all formatters..."
	@command -v gofmt >/dev/null 2>&1 && gofmt -w -s . || echo "gofmt not available"
	@command -v goimports >/dev/null 2>&1 && \
		goimports -w -local github.com/ivuorinen/gh-action-readme . || \
		echo "goimports not available"

editorconfig: ## Check EditorConfig compliance
	@echo "Checking EditorConfig compliance..."
	@command -v eclint >/dev/null 2>&1 || \
		{ echo "Please install eclint: npm install -g eclint"; exit 1; }
	@echo "Checking files for EditorConfig compliance..."
	@find . -type f \( \
		-name "*.go" -o \
		-name "*.yml" -o \
		-name "*.yaml" -o \
		-name "*.json" -o \
		-name "*.md" -o \
		-name "Makefile" -o \
		-name "*.tmpl" -o \
		-name "*.adoc" -o \
		-name "*.sh" \
	\) -not -path "./.*" -not -path "./gh-action-readme" -not -path "./coverage*" \
		-not -path "./testutil.test" -not -path "./test_*" | \
		xargs eclint check

editorconfig-fix: ## Fix EditorConfig violations
	@echo "Fixing EditorConfig violations..."
	@command -v eclint >/dev/null 2>&1 || \
		{ echo "Please install eclint: npm install -g eclint"; exit 1; }
	@echo "Fixing files for EditorConfig compliance..."
	@find . -type f \( \
		-name "*.go" -o \
		-name "*.yml" -o \
		-name "*.yaml" -o \
		-name "*.json" -o \
		-name "*.md" -o \
		-name "Makefile" -o \
		-name "*.tmpl" -o \
		-name "*.adoc" -o \
		-name "*.sh" \
	\) -not -path "./.*" -not -path "./gh-action-readme" -not -path "./coverage*" \
		-not -path "./testutil.test" -not -path "./test_*" | \
		xargs eclint fix

# Security targets
security: vulncheck snyk trivy gitleaks ## Run all security scans
	@echo "All security scans completed"

vulncheck: ## Run Go vulnerability check
	@echo "Running Go vulnerability check..."
	@command -v govulncheck >/dev/null 2>&1 || \
		{ echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	govulncheck ./...

audit: vulncheck ## Run comprehensive security audit
	@echo "Running comprehensive security audit..."
	go list -json -deps ./... | jq -r '.Module | select(.Path != null) | .Path + "@" + .Version' | sort -u

snyk: ## Run Snyk security scan
	@echo "Running Snyk security scan..."
	@command -v snyk >/dev/null 2>&1 || \
		{ echo "Please install Snyk CLI: npm install -g snyk"; exit 1; }
	snyk test --file=go.mod --package-manager=gomodules

trivy: ## Run Trivy filesystem scan
	@echo "Running Trivy filesystem scan..."
	@command -v trivy >/dev/null 2>&1 || \
		{ echo "Please install Trivy: https://aquasecurity.github.io/trivy/"; exit 1; }
	trivy fs . --severity HIGH,CRITICAL

gitleaks: ## Run gitleaks secrets detection
	@echo "Running gitleaks secrets detection..."
	@command -v gitleaks >/dev/null 2>&1 || \
		{ echo "Please install gitleaks: https://github.com/gitleaks/gitleaks"; exit 1; }
	gitleaks detect --source . --verbose
