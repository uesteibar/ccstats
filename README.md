> ⚠︎ This tool was built using an AI agent

# ccstats


A macOS CLI tool to display Claude Code usage statistics and Codex usage limits in one view.

## Features

- ASCII progress bars showing usage percentage
- Color-coded output based on usage levels (green < 50%, yellow 50-80%, red > 80%)
- Human-readable reset times
- Reuses existing OAuth credentials from macOS Keychain (no separate login required)
- TTY detection for automatic color disabling when piped
- Codex plan detection from `~/.codex/auth.json`
- Codex usage limits table for all plans

## Installation

```bash
go install github.com/uesteibar/ccstats@latest
```

## Prerequisites

- macOS (uses Keychain for credential storage)
- [Claude Code](https://claude.ai/claude-code) must be installed and authenticated
- Go 1.24+ (for installation from source)

## Usage

### Display Usage Statistics (Claude + Codex)

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

Codex Usage Limits (Plan: Plus)
────────────────────────────────────────────────────────────
7-day          [██████░░░░░░░░░░░░░░]  30%  resets in 4d 2h
5-hour         [████░░░░░░░░░░░░░░░░]  20%  resets in 2h 10m
```

### Display Codex Usage Limits

```bash
ccstats codex
```

Example output:

```
Codex Usage Limits (Plan: Plus)
────────────────────────────────────────────────────────────
7-day          [██████░░░░░░░░░░░░░░]  30%  resets in 4d 2h
5-hour         [████░░░░░░░░░░░░░░░░]  20%  resets in 2h 10m
```

### Check Authentication Status

```bash
ccstats auth
# or
ccstats status
```

This verifies if valid credentials are found in Keychain without making API calls.

To check Codex credentials:

```bash
ccstats codex auth
# or
ccstats codex status
```

## How It Works

`ccstats` reads OAuth credentials stored by Claude Code in your macOS Keychain and fetches usage data from Anthropic's API. No additional authentication is required if you're already logged into Claude Code.

If you see an authentication error, run `claude` in your terminal to authenticate.

For Codex limits, `ccstats` reads `~/.codex/auth.json` (or the `OPENAI_API_KEY` environment variable) to determine your plan.

## License

MIT
