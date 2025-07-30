#!/bin/bash
# Release script for gh-action-readme
# Usage: ./scripts/release.sh [version]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [[ ! -f ".goreleaser.yaml" ]]; then
  log_error "This script must be run from the project root directory"
  exit 1
fi

# Check if GoReleaser is installed
if ! command -v goreleaser &>/dev/null; then
  log_error "GoReleaser is not installed. Install it first:"
  echo "  brew install goreleaser/tap/goreleaser"
  echo "  or visit: https://goreleaser.com/install/"
  exit 1
fi

# Get version from command line or prompt
VERSION="$1"
if [[ -z "$VERSION" ]]; then
  echo -n "Enter version (e.g., v1.0.0): "
  read -r VERSION
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  log_error "Version must be in format vX.Y.Z (e.g., v1.0.0)"
  exit 1
fi

log_info "Preparing release $VERSION"

# Check if we're on main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
  log_warning "You're not on the main branch (current: $CURRENT_BRANCH)"
  echo -n "Continue anyway? (y/N): "
  read -r CONTINUE
  if [[ "$CONTINUE" != "y" && "$CONTINUE" != "Y" ]]; then
    log_info "Aborted"
    exit 0
  fi
fi

# Check for uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
  log_error "You have uncommitted changes. Please commit or stash them first."
  git status --short
  exit 1
fi

# Update CHANGELOG.md
log_info "Please update CHANGELOG.md with changes for $VERSION"
echo -n "Press Enter when ready to continue..."
read -r

# Run tests and linting
log_info "Running tests and linting..."
if ! go test ./...; then
  log_error "Tests failed. Please fix them before releasing."
  exit 1
fi

if ! golangci-lint run; then
  log_error "Linting failed. Please fix issues before releasing."
  exit 1
fi

# Build and test GoReleaser config
log_info "Testing GoReleaser configuration..."
if ! goreleaser check; then
  log_error "GoReleaser configuration is invalid"
  exit 1
fi

# Test build without releasing
log_info "Testing release build..."
if ! goreleaser build --snapshot --clean; then
  log_error "Release build failed"
  exit 1
fi

log_success "Build test completed successfully"

# Commit any pending changes (like CHANGELOG updates)
if [[ -n $(git status --porcelain) ]]; then
  log_info "Committing pending changes..."
  git add .
  git commit -m "chore: prepare release $VERSION"
fi

# Create and push tag
log_info "Creating and pushing tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

log_success "Tag $VERSION created and pushed"
log_info "GitHub Actions will now build and publish the release automatically"
log_info "Check the progress at: https://github.com/ivuorinen/gh-action-readme/actions"

# Open release page
if command -v open &>/dev/null; then
  log_info "Opening release page..."
  open "https://github.com/ivuorinen/gh-action-readme/releases/tag/$VERSION"
elif command -v xdg-open &>/dev/null; then
  log_info "Opening release page..."
  xdg-open "https://github.com/ivuorinen/gh-action-readme/releases/tag/$VERSION"
fi

log_success "Release process initiated for $VERSION"
