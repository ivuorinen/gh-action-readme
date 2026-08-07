#!/usr/bin/env python3
"""Detect inline YAML/config data embedded in Go test files.

Enforces .claude/rules/no-inline-yaml-in-tests.md. Shared by two callers so a
missed agent-hook invocation is still caught before a commit lands:

  * .claude/hooks/pre-test-inline-yaml-guard.sh  -- PreToolUse(Edit|Write), scans
    the pending content on stdin via --stdin-content.
  * .pre-commit-config.yaml                      -- scans staged files by path.

Exit status: 0 clean, 1 violation found (message on stdout).

The heuristic looks for a Go string literal (raw backtick or double-quoted with
escapes decoded) containing two or more YAML key lines. Nesting is NOT required:
flat metadata such as "name: X\\ndescription: Y\\n" is exactly the inline fixture
the rule forbids, and requiring an indented line let it through.
"""

from __future__ import annotations

import argparse
import json
import re
import sys

# A YAML key line: `word:` followed by a space or end-of-line. Requiring the
# space/EOL keeps ordinary prose containing a colon ("note: see above" inside a
# sentence is still a key, but "http://x" and "Errorf(%q): %v" are not) from
# matching, which is what keeps the false-positive rate low.
KEY = re.compile(r"^[ \t]*[A-Za-z0-9_.-]+:(?: |$)")

MIN_KEY_LINES = 2

# A "key" whose value is only a printf verb is a Go format string ("stdout: %s\n
# stderr: %s"), not YAML. Excluding these is what lets the indentation requirement
# be dropped without drowning in false positives.
FORMAT_VALUE = re.compile(r"^[ \t]*[A-Za-z0-9_.-]+:\s*%[-+ #0-9.]*[a-zA-Z]\s*$")

# Go source accidentally captured between two struct-tag backticks. A YAML document
# has none of these, and a Go struct literal or func body has at least one.
GO_TOKEN = re.compile(r"(\bfunc\b|:=|\*testing\.T|\{\s*$|^\s*\}|,\s*$)")

MESSAGE = (
    "inline YAML in a *_test.go file (no-inline-yaml-in-tests.md). "
    "Move it to testdata/yaml-fixtures/, load it via testutil, and register the "
    "path in testutil/test_constants.go."
)


def string_literals(source: str) -> list[str]:
    """Return the contents of every Go string literal in source.

    Raw (backtick) literals are taken verbatim. Double-quoted literals have their
    escapes decoded, because inline YAML written as "a: b\\n  c: d\\n" is the
    common form and a backtick-only scan misses it entirely.
    """
    blocks = re.findall(r"`([^`]*)`", source, re.S)

    for quoted in re.findall(r'"((?:[^"\\]|\\.)*)"', source):
        try:
            blocks.append(quoted.encode("latin-1", "ignore").decode("unicode_escape"))
        except (UnicodeDecodeError, UnicodeEncodeError):
            continue

    return blocks


def block_is_yaml(block: str) -> bool:
    """Report whether one string literal's contents look like a YAML document."""
    lines = block.splitlines()

    # Go source captured between two struct-tag backticks is not a YAML document.
    if any(GO_TOKEN.search(line) for line in lines):
        return False

    keyed = [line for line in lines if KEY.match(line) and not FORMAT_VALUE.match(line)]

    return len(keyed) >= MIN_KEY_LINES


def has_inline_yaml(source: str) -> bool:
    """Report whether source embeds a YAML-shaped document in a string literal."""
    return any(block_is_yaml(block) for block in string_literals(source))


def read_stdin_content() -> tuple[str, str]:
    """Parse a Claude Code hook payload from stdin into (path, pending content)."""
    try:
        payload = json.loads(sys.stdin.read() or "{}")
    except (ValueError, TypeError):
        return "", ""  # fail open: never block on a malformed payload

    tool_input = payload.get("tool_input") or {}
    path = tool_input.get("file_path") or ""
    # Edit supplies new_string; Write supplies content.
    content = tool_input.get("new_string") or tool_input.get("content") or ""

    return path, content


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--stdin-content",
        action="store_true",
        help="read a Claude Code hook payload from stdin instead of scanning paths",
    )
    parser.add_argument("paths", nargs="*", help="files to scan")
    args = parser.parse_args()

    if args.stdin_content:
        path, content = read_stdin_content()
        if not path.endswith("_test.go") or not content:
            return 0
        if has_inline_yaml(content):
            print(f"BLOCKED: {MESSAGE}")

            return 1

        return 0

    failed = False
    for path in args.paths:
        if not path.endswith("_test.go"):
            continue
        try:
            with open(path, encoding="utf-8") as handle:
                source = handle.read()
        except OSError as err:
            print(f"{path}: cannot read ({err})")
            failed = True

            continue

        if has_inline_yaml(source):
            print(f"{path}: {MESSAGE}")
            failed = True

    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
