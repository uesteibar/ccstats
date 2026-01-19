> ⚠︎ This tool was built using an AI agent

# ccstats


A macOS CLI tool to display Claude Code usage statistics with visual progress bars and color-coded output.

## Features

- ASCII progress bars showing usage percentage
- Color-coded output based on usage levels (green < 50%, yellow 50-80%, red > 80%)
- Human-readable reset times
- Reuses existing OAuth credentials from macOS Keychain (no separate login required)
- TTY detection for automatic color disabling when piped

## Installation

```bash
go install github.com/uesteibar/ccstats@latest
```

## Prerequisites

- macOS (uses Keychain for credential storage)
- [Claude Code](https://claude.ai/claude-code) must be installed and authenticated
- Go 1.24+ (for installation from source)

## Usage

### Display Usage Statistics

```bash
ccstats
```

Example output:

```
Claude Code Usage Statistics
────────────────────────────────────────────────────────────
5-hour         [████████░░░░░░░░░░░░]  40%  resets in 2h 15m
7-day          [██████████████░░░░░░]  70%  resets in 3d 5h
7-day Opus     [██░░░░░░░░░░░░░░░░░░]  10%  resets in 3d 5h
```

### Check Authentication Status

```bash
ccstats auth
# or
ccstats status
```

This verifies if valid credentials are found in Keychain without making API calls.

## How It Works

`ccstats` reads OAuth credentials stored by Claude Code in your macOS Keychain and fetches usage data from Anthropic's API. No additional authentication is required if you're already logged into Claude Code.

If you see an authentication error, run `claude` in your terminal to authenticate.

## License

MIT
