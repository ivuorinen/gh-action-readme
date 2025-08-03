.PHONY: test lint run example clean readme config-verify security vulncheck audit snyk trivy gitleaks \
	editorconfig editorconfig-fix format

all: test lint

test:
	go test ./...

lint: editorconfig
	golangci-lint run || true

config-verify:
	golangci-lint config verify --verbose

run:
	go run .

example:
	go run . gen --config config.yaml --output-format=md

readme:
	go run . gen --config config.yaml --output-format=md

clean:
	rm -rf dist/

# Code formatting and EditorConfig targets
format: editorconfig-fix
	@echo "Running all formatters..."
	@command -v gofmt >/dev/null 2>&1 && gofmt -w -s . || echo "gofmt not available"
	@command -v goimports >/dev/null 2>&1 && \
		goimports -w -local github.com/ivuorinen/gh-action-readme . || \
		echo "goimports not available"

editorconfig:
	@echo "Checking EditorConfig compliance..."
	@command -v eclint >/dev/null 2>&1 || \
		{ echo "Please install eclint: npm install -g eclint"; exit 1; }
	@echo "Checking key files for EditorConfig compliance..."
	eclint check Makefile .github/workflows/*.yml main.go internal/**/*.go *.md .goreleaser.yaml

editorconfig-fix:
	@echo "Fixing EditorConfig violations..."
	@command -v eclint >/dev/null 2>&1 || \
		{ echo "Please install eclint: npm install -g eclint"; exit 1; }
	@echo "Fixing key files for EditorConfig compliance..."
	eclint fix Makefile .github/workflows/*.yml main.go internal/**/*.go *.md .goreleaser.yaml

editorconfig-check:
	@echo "Running editorconfig-checker..."
	@command -v editorconfig-checker >/dev/null 2>&1 || \
		{ echo "Please install editorconfig-checker: npm install -g editorconfig-checker"; exit 1; }
	editorconfig-checker

# Security targets
security: vulncheck snyk trivy gitleaks
	@echo "All security scans completed"

vulncheck:
	@echo "Running Go vulnerability check..."
	@command -v govulncheck >/dev/null 2>&1 || \
		{ echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	govulncheck ./...

audit: vulncheck
	@echo "Running comprehensive security audit..."
	go list -json -deps ./... | jq -r '.Module | select(.Path != null) | .Path + "@" + .Version' | sort -u

snyk:
	@echo "Running Snyk security scan..."
	@command -v snyk >/dev/null 2>&1 || \
		{ echo "Please install Snyk CLI: npm install -g snyk"; exit 1; }
	snyk test --file=go.mod --package-manager=gomodules

trivy:
	@echo "Running Trivy filesystem scan..."
	@command -v trivy >/dev/null 2>&1 || \
		{ echo "Please install Trivy: https://aquasecurity.github.io/trivy/"; exit 1; }
	trivy fs . --severity HIGH,CRITICAL

gitleaks:
	@echo "Running gitleaks secrets detection..."
	@command -v gitleaks >/dev/null 2>&1 || \
		{ echo "Please install gitleaks: https://github.com/gitleaks/gitleaks"; exit 1; }
	gitleaks detect --source . --verbose
