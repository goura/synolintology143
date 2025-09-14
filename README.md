# synolintology143
Finds the one weird filename that's about to silently ruin your synology backup vibe

## Motivation

I regularly backup from Linux to Synology's eCryptfs and it often fails because the file length is too long so I made a script to run before I run rsnapshot in cron

## Usage

Run it against one or more directories. It's read-only and won't change a thing.

Flags:
- `-q`, `--quiet` — silence stderr (progress and friendly messages). Stdout behavior is unchanged.

### Scan two shares for any problematic filenames
synolintology143 /share/Photos /share/Documents


#### If everything is chill:
It writes a friendly note to stderr and exits with code 0. Stdout stays empty.

stderr:
```
--- Scanning: /share/Photos ---
--- Scanning: /share/Documents ---

All good: no violating filenames found.
```

#### If it finds a problem:
It prints violating file paths to stdout (newline-separated, one per line) and exits with code 1. Helpful info goes to stderr.

stderr:
```
--- Scanning: /share/FamilyPhotosRecent ---

Heads up: violating filenames were found.
```

stdout:
```
/share/FamilyPhotosRecent/2025/2025-03-28/2025-03-28_17-16-50_..._fe3c25d17477.webp
```

This makes it easy to pipe into other tools. For example, to preview the list:
```
synolintology143 /share/Photos | sed -n '1,20p'
```

Or to act on the results (use with care):
```
synolintology143 /share/Photos | while read -r p; do echo "would handle: $p"; done
```

Exit codes:
- 0: no violating filenames found (stderr may contain friendly info)
- 1: one or more violating filenames printed to stdout
- 2: usage error (bad/missing args)

Quiet examples:
```
# No progress or summary on stderr, only stdout if there are violations
synolintology143 -q /share/Photos | wc -l
synolintology143 --quiet /share/Photos /share/Documents > offenders.txt
```

## Dev: Formatting Hooks

Use either the pre-commit framework or a plain Git hook to enforce `gofmt`.

- Pre-commit (recommended):
  - Install: `pipx install pre-commit` or `pip install pre-commit`
  - Enable: `pre-commit install`
  - Run on all files: `pre-commit run --all-files`
  - Config: see `.pre-commit-config.yaml` (includes a `gofmt` check and a manual `gofmt -w` fixer).

- Plain Git hook:
  - Set hooks path: `git config core.hooksPath scripts/githooks`
  - Make it executable: `chmod +x scripts/githooks/pre-commit`
  - Now commits will fail if any staged `.go` files aren’t `gofmt`-clean.
