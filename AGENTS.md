# AGENTS.md

This file guides agents and contributors working in this repository. It applies to the entire repo.

## Project Overview
- Purpose: Find filenames that would exceed eCryptfs’ effective 143‑byte plaintext name limit and break Synology backups.
- Language/Tooling: Go (module `github.com/goura/synolintology143`, `go 1.25.1`). No external deps beyond the standard library.

## Output Contract (Do Not Break)
Strict rules for CLI behavior. Any change must keep these guarantees and update tests/README if needed.

- Stdout
  - Default mode: print a newline‑separated list of violating file paths only. Print nothing to stdout when there are no violations. Each violating path ends with a single newline. No extra commentary.
  - JSON mode (`-j`/`--json`): print a single JSON array of violating file paths to stdout. On success, print `[]` (valid JSON). A trailing newline after the JSON value is acceptable.
- Stderr
  - Progress banners: `--- Scanning: <path> ---`.
  - Warnings and helpful notes live here.
  - On success (no violations): print a short, calm summary to stderr.
  - On failure (violations present): optional short note to stderr; do not duplicate paths here.
- Exit codes
  - `0`: no violations (stdout empty, stderr may have a summary).
  - `1`: violations found (stdout has paths).
  - `2`: usage error (e.g., missing args).
- Quiet flag
  - `-q` / `--quiet` silences stderr (progress, warnings, summaries). Stdout behavior unchanged (newline list in default mode; JSON array in JSON mode).

## Code Structure
- `cmd/synolintology143/main.go`
  - Parses flags/args (`-q/--quiet`, `-j/--json`).
  - Wires scanner writers (silences `ErrWriter` when quiet).
  - Handles exit codes and minimal stderr notes.
- `internal/scanner/scanner.go`
  - Core walk/scan logic and limit constant (`MAX_ECRYPTFS_FILENAME_BYTES = 143`).
  - Writers: `OutWriter` (defaults to `os.Stdout`), `ErrWriter` (defaults to `os.Stderr`).
  - Print violating paths with `fmt.Fprintln(OutWriter, path)`; log info/warnings to `ErrWriter`.
  - JSON mode is implemented by wrapping `OutWriter` in the CLI with a writer that formats newline events into a JSON array; the scanner remains unchanged.

## Tests
- Location: `internal/scanner/scanner_test.go`.
- Strategy
  - Capture `scanner.OutWriter` via an `os.Pipe()` to assert stdout exactly.
  - Set `scanner.ErrWriter = io.Discard` in tests to avoid progress noise.
  - Verify: no stdout for clean trees; newline‑separated paths for violations; no good files in output.
- JSON tests live in `cmd/synolintology143/main_test.go` and exercise the JSON writer by wrapping `scanner.OutWriter`.
- Run: `go test ./...`

## Development Workflow
- Build: `go build ./cmd/synolintology143`
- Lint/format: rely on `gofmt`/`go vet` defaults. Avoid adding new tooling.
- Dependencies: prefer standard library only.
- Changes should be minimal and focused. If you alter user‑visible behavior, update:
  1) tests, 2) README examples, 3) this AGENTS.md if the contract changes.

## Conventions
- Keep CLI messages short and friendly. No colors or emojis.
- Avoid adding new flags unless they simplify scripting and keep the contract intact.
- Do not emit JSON or structured formats unless explicitly requested and documented.

## Token‑Saving Tips for Agents
- Key files to inspect first:
  - `internal/scanner/scanner.go`, `internal/scanner/scanner_test.go`, `cmd/synolintology143/main.go`, `README.md`.
- Common queries:
  - “Where is stdout printed?” → search for `OutWriter` and `Fprintln`.
  - “Where is stderr printed?” → search for `ErrWriter`, banners, and warnings.
- Use fast searches: `rg -n "OutWriter|ErrWriter|fmt\.(Print|Fprint)|Scanning:"`.
- Keep chat replies concise. Group shell reads into a single preamble, per step.
- Only add a plan for multi‑step tasks; otherwise, do the change directly.

## Non‑Goals
- Do not change the output format or exit codes without an explicit request.
- Do not introduce runtime dependencies or logging frameworks.

## Updating Limits (If Ever Needed)
- Byte limit is hard‑coded at 143. If the requirement changes, update:
  - `internal/scanner/scanner.go` constant, tests assertions, README, and this file.
