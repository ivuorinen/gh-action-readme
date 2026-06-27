# Context-Mode Tools Required

Always run context-producing commands (e.g. `go test`, builds, large output) through the
context-mode tool `mcp__plugin_context-mode_context-mode__ctx_execute` so output stays in the
sandbox and only a summary enters context.

Never bypass it with the `# ctx-ok` Bash escape just to avoid context-mode. The escape exists
only for genuinely non-context-producing commands.

When running a `/goal` (or any goal-driven task): if the context-mode tools are not available in
the session, treat the task as **failed** and stop. Do not fall back to `# ctx-ok` or raw Bash to
keep going. Report the unavailability as the failure reason.
