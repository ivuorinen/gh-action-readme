# Dockerfile for gh-action-readme
FROM scratch

# Multi-platform build support
# See: https://goreleaser.com/customization/dockers_v2/
# GoReleaser organizes binaries in platform subdirectories (linux/amd64/, linux/arm64/)
# TARGETPLATFORM arg resolves to the correct platform directory
ARG TARGETPLATFORM

# Copy the binary from the build context (platform-specific)
COPY $TARGETPLATFORM/gh-action-readme /usr/local/bin/gh-action-readme

# Copy schemas (templates are embedded in the binary via go:embed)
COPY schemas /usr/local/share/gh-action-readme/schemas

# Set environment variable for schema path
ENV GH_ACTION_README_SCHEMA_PATH=/usr/local/share/gh-action-readme/schemas

# Run as non-root (numeric UID required for scratch/distroless images with no /etc/passwd)
USER 65532:65532

# No health check — this is a CLI tool, not a long-running service
HEALTHCHECK NONE

# Set the binary as entrypoint
ENTRYPOINT ["/usr/local/bin/gh-action-readme"]

# Default command
CMD ["--help"]

# Labels for metadata
LABEL org.opencontainers.image.title="gh-action-readme"
LABEL org.opencontainers.image.description="Auto-generate beautiful README and HTML documentation for GitHub Actions"
LABEL org.opencontainers.image.url="https://github.com/ivuorinen/gh-action-readme"
LABEL org.opencontainers.image.source="https://github.com/ivuorinen/gh-action-readme"
LABEL org.opencontainers.image.vendor="ivuorinen"
LABEL org.opencontainers.image.licenses="MIT"
