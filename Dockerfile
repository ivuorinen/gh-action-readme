# Dockerfile for gh-action-readme
FROM scratch

# Copy the binary from the build context
COPY gh-action-readme /usr/local/bin/gh-action-readme

# Copy templates and schemas
COPY templates /usr/local/share/gh-action-readme/templates
COPY schemas /usr/local/share/gh-action-readme/schemas

# Set environment variables for template paths
ENV GH_ACTION_README_TEMPLATE_PATH=/usr/local/share/gh-action-readme/templates
ENV GH_ACTION_README_SCHEMA_PATH=/usr/local/share/gh-action-readme/schemas

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
