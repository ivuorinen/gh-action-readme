.PHONY: help test test-coverage test-coverage-html lint build run example \
	clean readme config-verify security vulncheck audit snyk trivy gitleaks \
	editorconfig editorconfig-fix format devtools pre-commit-install pre-commit-update

all: help

help: ## Show this help message
	@echo "GitHub Action README Generator - Available Make Targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Common workflows:"
	@echo "  make devtools            # Install all development tools"
	@echo "  make pre-commit-install  # Install pre-commit hooks (run once)"
	@echo "  make build               # Build the application binary"
	@echo "  make test lint           # Run tests and all linters via pre-commit"
	@echo "  make test-coverage       # Run tests with coverage analysis"
	@echo "  make pre-commit-update   # Update pre-commit hooks to latest versions"
	@echo "  make security            # Run all security scans"

test: ## Run all tests
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

lint: ## Run all linters via pre-commit
	@echo "Running all linters via pre-commit..."
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Please install pre-commit or run 'make devtools'"; exit 1; }
	pre-commit run --all-files

pre-commit-install: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Please install pre-commit or run 'make devtools'"; exit 1; }
	pre-commit install

pre-commit-update: ## Update pre-commit hooks to latest versions
	@echo "Updating pre-commit hooks..."
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Please install pre-commit or run 'make devtools'"; exit 1; }
	pre-commit autoupdate

build: ## Build the application
	go build -o gh-action-readme .

config-verify: ## Verify golangci-lint configuration
	golangci-lint config verify --verbose

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
	@command -v gofmt >/dev/null 2>&1 && gofmt -w -s . || echo "gofmt not available"
	@command -v goimports >/dev/null 2>&1 && \
		goimports -w -local github.com/ivuorinen/gh-action-readme . || \
		echo "goimports not available"

editorconfig: ## Check EditorConfig compliance
	@echo "Checking EditorConfig compliance..."
	@command -v editorconfig-checker >/dev/null 2>&1 || \
		{ echo "Please install editorconfig-checker or run 'make devtools'"; exit 1; }
	editorconfig-checker

editorconfig-fix: ## Fix EditorConfig violations
	@echo "EditorConfig violations cannot be automatically fixed by editorconfig-checker"
	@echo "Please fix the reported issues manually or use your editor's EditorConfig plugin"
	@echo "Running check to show issues..."
	@command -v editorconfig-checker >/dev/null 2>&1 || \
		{ echo "Please install editorconfig-checker or run 'make devtools'"; exit 1; }
	editorconfig-checker

# Development tools installation
devtools: ## Install all development tools
	@echo "Installing development tools..."
	@echo ""
	@echo "=== Go Tools ==="
	@command -v golangci-lint >/dev/null 2>&1 || \
		{ echo "Installing golangci-lint..."; \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
			sh -s -- -b $(go env GOPATH)/bin; }
	@command -v govulncheck >/dev/null 2>&1 || \
		{ echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	@command -v editorconfig-checker >/dev/null 2>&1 || \
		{ echo "Installing editorconfig-checker..."; \
			go install github.com/editorconfig-checker/editorconfig-checker/v3/cmd/editorconfig-checker@latest; }
	@command -v yamlfmt >/dev/null 2>&1 || \
		{ echo "Installing yamlfmt..."; go install github.com/google/yamlfmt/cmd/yamlfmt@latest; }
	@echo "✓ Go tools installed"
	@echo ""
	@echo "=== Node.js Tools ==="
	@command -v npm >/dev/null 2>&1 || \
		{ echo "❌ npm not found. Please install Node.js first."; exit 1; }
	@command -v snyk >/dev/null 2>&1 || \
		{ echo "Installing snyk..."; npm install -g snyk; }
	@echo "✓ Node.js tools installed"
	@echo ""
	@echo "=== Python Tools ==="
	@command -v python3 >/dev/null 2>&1 || \
		{ echo "❌ python3 not found. Please install Python 3 first."; exit 1; }
	@command -v pre-commit >/dev/null 2>&1 || \
		{ echo "Installing pre-commit..."; pip install pre-commit; }
	@echo "✓ Python tools installed"
	@echo ""
	@echo "=== System Tools ==="
	@command -v trivy >/dev/null 2>&1 || \
		{ echo "❌ trivy not found. Please install manually: https://aquasecurity.github.io/trivy/"; }
	@command -v gitleaks >/dev/null 2>&1 || \
		{ echo "❌ gitleaks not found. Please install manually: https://github.com/gitleaks/gitleaks"; }
	@echo "✓ System tools check completed"
	@echo ""
	@echo "🎉 Development tools installation completed!"
	@echo "   Run 'make test lint' to verify everything works."

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
