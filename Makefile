.PHONY: help test test-quick test-coverage test-coverage-html test-coverage-check \
	test-mutation test-mutation-parser test-mutation-validation \
	test-property test-property-validation test-property-parser \
	lint build run example clean readme config-verify \
	security vulncheck gosec audit trivy gitleaks \
	editorconfig editorconfig-fix format format-yaml devtools \
	pre-commit-install pre-commit-update \
	deps-check deps-update deps-update-all

all: help

# Coverage threshold (align with SonarCloud)
# Note: SonarCloud checks NEW code coverage (≥80%), this checks overall coverage
# Current overall coverage: 73.7% - working towards 80% target
COVERAGE_THRESHOLD := 72.0

# Tool versions (managed by Renovate)
# renovate: datasource=go depName=golang.org/x/vuln
GOVULNCHECK_VERSION := v1.5.0
# renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
GOLANGCI_LINT_VERSION := v2.12.2
# renovate: datasource=go depName=github.com/editorconfig-checker/editorconfig-checker/v3
EDITORCONFIG_CHECKER_VERSION := v3.7.0
# renovate: datasource=go depName=github.com/oligot/go-mod-upgrade
GO_MOD_UPGRADE_VERSION := v0.12.0
# renovate: datasource=go depName=github.com/securego/gosec/v2
GOSEC_VERSION := v2.27.1
# renovate: datasource=go depName=github.com/google/yamlfmt
YAMLFMT_VERSION := v0.21.0
# renovate: datasource=go depName=golang.org/x/tools
GOIMPORTS_VERSION := v0.47.0
# renovate: datasource=go depName=github.com/go-gremlins/gremlins
GREMLINS_VERSION := v0.6.0

# Tool command shortcuts (avoids long lines in recipes)
EC_MODULE := github.com/editorconfig-checker/editorconfig-checker/v3/cmd/editorconfig-checker
GOLANGCI_MODULE := github.com/golangci/golangci-lint/v2/cmd/golangci-lint
GOVULNCHECK_MODULE := golang.org/x/vuln/cmd/govulncheck
GOSEC_MODULE := github.com/securego/gosec/v2/cmd/gosec
YAMLFMT_MODULE := github.com/google/yamlfmt/cmd/yamlfmt
GOIMPORTS_MODULE := golang.org/x/tools/cmd/goimports
GO_MOD_UPGRADE_MODULE := github.com/oligot/go-mod-upgrade
GREMLINS_MODULE := github.com/go-gremlins/gremlins/cmd/gremlins

EC := go run $(EC_MODULE)@$(EDITORCONFIG_CHECKER_VERSION)
GOLANGCI := go run $(GOLANGCI_MODULE)@$(GOLANGCI_LINT_VERSION)
GOVULNCHECK := go run $(GOVULNCHECK_MODULE)@$(GOVULNCHECK_VERSION)
GOSEC := go run $(GOSEC_MODULE)@$(GOSEC_VERSION)
YAMLFMT := go run $(YAMLFMT_MODULE)@$(YAMLFMT_VERSION)
GOIMPORTS := go run $(GOIMPORTS_MODULE)@$(GOIMPORTS_VERSION)
GO_MOD_UPGRADE := go run $(GO_MOD_UPGRADE_MODULE)@$(GO_MOD_UPGRADE_VERSION)
GREMLINS := go run $(GREMLINS_MODULE)@$(GREMLINS_VERSION)

help: ## Show this help message
	@echo "GitHub Action README Generator - Available Make Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Common workflows:"
	@echo "  make devtools            # Install system-only tools (trivy, gitleaks)"
	@echo "  make pre-commit-install  # Install pre-commit hooks (run once)"
	@echo "  make build               # Build the application binary"
	@echo "  make test lint           # Run tests and all linters via pre-commit"
	@echo "  make test-coverage       # Run tests with coverage analysis"
	@echo "  make pre-commit-update   # Update pre-commit hooks to latest versions"
	@echo "  make deps-check          # Check for outdated dependencies"
	@echo "  make deps-update         # Update dependencies interactively"
	@echo "  make security            # Run all security scans"

test: ## Run all tests (standard and property-based)
	@echo "Running standard tests..."
	@go test ./...
	@echo ""
	@echo "Running property-based tests..."
	@$(MAKE) test-property
	@echo ""
	@echo "✅ All tests (standard + property) completed successfully!"
	@echo ""
	@echo "Note: Run 'make test-mutation' to execute mutation tests via gremlins."
	@echo "      Run 'make test-quick' for fast iteration (unit tests only)."

test-quick: ## Run only standard unit tests (fast)
	go test ./...

test-coverage: ## Run tests with coverage and display in CLI
	@echo "Running tests with coverage analysis..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@echo ""
	@echo "=== Coverage Summary ==="
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "=== Package Coverage Details ==="
	@go tool cover -func=coverage.out | grep -v "total:" | \
		awk '{printf "%-50s %s\n", $$1, $$3}' | \
		sort -k2 -nr
	@echo ""
	@echo "Coverage report saved to: coverage.out"
	@echo "Run 'make test-coverage-html' to generate HTML report"

test-coverage-html: test-coverage ## Generate HTML coverage report and open in browser
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated: coverage.html"
	@if command -v open >/dev/null 2>&1; then \
		echo "Opening coverage report in browser..."; \
		open coverage.html; \
	elif command -v xdg-open >/dev/null 2>&1; then \
		echo "Opening coverage report in browser..."; \
		xdg-open coverage.html; \
	else \
		echo "Open coverage.html in your browser to view detailed coverage"; \
	fi

test-coverage-check: ## Run tests with coverage check (overall >= 72%)
	@command -v bc >/dev/null 2>&1 || { \
		echo "❌ bc command not found. Please install bc (e.g., apt-get install bc, brew install bc)"; \
		exit 1; \
	}
	@echo "Running tests with coverage check..."
	@go test -cover -coverprofile=coverage.out ./...
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc) -eq 1 ]; then \
		echo "❌ Coverage $$total% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$total% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

.PHONY: test-mutation test-mutation-parser test-mutation-validation

test-mutation: test-mutation-parser test-mutation-validation ## Run all mutation tests

test-mutation-parser: ## Run mutation tests on parser (permission parsing)
	@echo "Running mutation tests on parser (gremlins)..."
	$(GREMLINS) unleash --timeout-coefficient 10 --workers 1 ./internal

test-mutation-validation: ## Run mutation tests on validation (version and strings)
	@echo "Running mutation tests on validation (gremlins)..."
	$(GREMLINS) unleash --timeout-coefficient 20 --workers 1 ./internal/validation

.PHONY: test-property test-property-validation test-property-parser

test-property: test-property-validation test-property-parser ## Run all property-based tests

test-property-validation: ## Run property tests on validation (strings)
	@echo "Running property tests on validation..."
	@go test -v ./internal/validation -run ".*Properties" -timeout 30s

test-property-parser: ## Run property tests on parser (permission merging)
	@echo "Running property tests on parser..."
	@go test -v ./internal -run ".*Properties" -timeout 30s

lint: editorconfig ## Run all linters via pre-commit
	@echo "Running all linters via pre-commit..."
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Please install pre-commit or run 'make devtools'"; exit 1; }
	pre-commit run --all-files

pre-commit-install: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Please install pre-commit or run 'make devtools'"; exit 1; }
	pre-commit install
	pre-commit install --hook-type commit-msg

pre-commit-update: ## Update pre-commit hooks to latest versions
	@echo "Updating pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Please install pre-commit or run 'make devtools'"; exit 1; }
	pre-commit autoupdate

build: ## Build the application
	go build -o gh-action-readme .

config-verify: ## Verify golangci-lint configuration
	$(GOLANGCI) config verify --verbose

run: ## Run the application
	go run .

example: ## Generate example README
	go run . gen --config config.yml --output-format=md

readme: ## Generate project README
	go run . gen --config config.yml --output-format=md

clean: ## Clean build artifacts
	rm -rf dist/
	rm -f gh-action-readme coverage.out coverage.html

# Code formatting and EditorConfig targets
format: editorconfig-fix ## Format code and fix EditorConfig issues
	@echo "Running all formatters..."
	gofmt -w -s .
	$(GOIMPORTS) -w -local github.com/ivuorinen/gh-action-readme .

format-yaml: ## Format YAML files
	@echo "Formatting YAML files..."
	$(YAMLFMT) .

editorconfig: ## Check EditorConfig compliance
	@echo "Checking EditorConfig compliance..."
	$(EC)

editorconfig-fix: ## Fix EditorConfig violations
	@echo "EditorConfig violations cannot be automatically fixed by editorconfig-checker"
	@echo "Please fix the reported issues manually or use your editor's EditorConfig plugin"
	@echo "Running check to show issues..."
	$(EC) || true

# Development tools installation (Go tools run via 'go run' — no install needed)
devtools: ## Install system-only development tools (trivy, gitleaks, pre-commit)
	@echo "Installing development tools..."
	@echo ""
	@echo "=== System Tools ==="
	@command -v trivy >/dev/null 2>&1 || \
		{ echo "❌ trivy not found. Please install manually: https://aquasecurity.github.io/trivy/"; }
	@command -v gitleaks >/dev/null 2>&1 || \
		{ echo "❌ gitleaks not found. Please install manually: https://github.com/gitleaks/gitleaks"; }
	@echo "✓ System tools check completed"
	@echo ""
	@echo "=== Python Tools ==="
	@command -v python3 >/dev/null 2>&1 || \
		{ echo "❌ python3 not found. Please install Python 3 first."; exit 1; }
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Installing pre-commit..."; pip install pre-commit; }
	@echo "✓ Python tools installed"
	@echo ""
	@echo "✅ Development tools setup completed!"
	@echo "   Go tools (golangci-lint, govulncheck, etc.) run via 'go run' automatically."
	@echo "   Run 'make test lint' to verify everything works."

# Security targets
security: vulncheck gosec trivy gitleaks ## Run all security scans
	@echo "All security scans completed"

vulncheck: ## Run Go vulnerability check
	@echo "Running Go vulnerability check..."
	$(GOVULNCHECK) ./...

gosec: ## Run gosec security scanner
	@echo "Running gosec security scanner..."
	$(GOSEC) ./...

audit: trivy gitleaks vulncheck gosec ## Run comprehensive security audit
	@echo "Running comprehensive security audit..."
	go list -json -deps ./... | jq -r '.Module | select(.Path != null) | .Path + "@" + .Version' | sort -u

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

# Dependency management targets
deps-check: ## Show outdated dependencies
	@echo "Checking for outdated dependencies..."
	@go list -u -m all | grep -v "^go: finding"

deps-update: ## Update dependencies interactively
	@echo "Starting interactive dependency update..."
	$(GO_MOD_UPGRADE)
	@echo "Running go mod tidy..."
	go mod tidy

deps-update-all: ## Update all dependencies to latest versions
	@echo "Updating all dependencies to latest versions..."
	@go get -u ./...
	@echo "Running go mod tidy..."
	go mod tidy
	@echo "All dependencies updated"
