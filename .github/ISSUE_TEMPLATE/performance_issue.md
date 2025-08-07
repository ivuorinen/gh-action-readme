---
name: Performance issue
about: Report CLI performance problems or slow operations
title: '[PERF] '
labels: performance, bug
assignees: ivuorinen

---

**Performance problem**
Describe what operation is slower than expected.

**Operation details**
Which gh-action-readme operation is slow?

- [ ] `gen` - Single file generation
- [ ] `gen --recursive` - Batch processing
- [ ] `validate` - Action validation
- [ ] `deps` - Dependency analysis
- [ ] `config` - Configuration operations
- [ ] Startup time
- [ ] Other: ___________

**Command executed**
The exact command that exhibits performance issues:

```bash
gh-action-readme [your slow command here]
```

**Performance metrics**
Provide timing and resource usage information:

**Processing time:**

- Current time: ___ seconds/minutes
- Expected time: ___ seconds/minutes
- Acceptable time: ___ seconds/minutes

**Dataset size:**

- Number of action.yml files: ___
- Total size of files: ___ KB/MB
- Repository structure depth: ___ levels
- Largest action.yml size: ___ KB

**Resource usage observed:**

- Peak memory usage: ___ MB
- CPU usage: ___% sustained
- Disk I/O patterns: [heavy reads, heavy writes, mixed]
- Network requests: ___ (for dependency analysis)

## Environment information

- OS: [e.g. macOS 14.1, Ubuntu 22.04, Windows 11]
- Hardware: [e.g. MacBook Air M2, Intel i7, AWS EC2 t3.large]
- gh-action-readme version: [run `gh-action-readme version`]
- Go version: [run `go version`]
- Installation method: [binary, homebrew, go install, docker]

**Batch processing details (if applicable)**
For recursive or batch operations:

**Repository structure:**

```text
my-repo/
├── .github/workflows/ (__ files)
├── actions/
│   ├── action1/ (action.yml)
│   ├── action2/ (action.yml)
│   └── ... (__ more actions)
└── other directories...
```

**Processing pattern:**

- [ ] Single large repository
- [ ] Multiple small repositories
- [ ] Mixed sizes
- [ ] Deep directory nesting
- [ ] Many small action.yml files
- [ ] Few large action.yml files

**Configuration impact**
Settings that might affect performance:

**Flags used:**

- Theme: [github, gitlab, minimal, professional, default]
- Output format: [md, html, json, asciidoc]
- Verbose mode: [yes, no]
- Dependency analysis: [enabled, disabled]

**Configuration file settings:**

```yaml
# Paste relevant config that might impact performance
```

## Expected vs actual behavior

- **Expected**: Should process files in specified seconds
- **Actual**: Takes ___ seconds/minutes to complete
- **Comparison**: Other similar tools take ___ seconds

**Profiling data (if available)**
If you've run any profiling:

```text
# Paste CPU/memory profiling output
# Or performance monitoring results
```

**Workarounds**
Any workarounds you've found:

- Breaking into smaller batches
- Specific flag combinations
- Environment modifications

## Additional context

- Network conditions (for dependency analysis)
- Disk type (SSD, HDD, network storage)
- Concurrent operations running
- Time of day patterns (if applicable)
