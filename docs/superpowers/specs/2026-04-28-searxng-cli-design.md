# SearXNG CLI - Design Document

## Overview

A single-binary Go CLI tool that queries SearXNG instances via their HTTP JSON API and outputs results in LLM-friendly compact JSON Lines format. The tool is configured via `~/.searxng_cli/config.yaml`.

## Goals

- Single binary, zero runtime dependencies
- LLM-friendly default output (compact JSONL, minimize token usage)
- Fast startup (< 100ms)
- Configurable SearXNG endpoint(s)

## Architecture

```
main.go → cmd/ (CLI parsing via cobra)
               ├── search.go   → client/ (HTTP API call) → formatter/ (output)
               └── config.go   → config/ (YAML read/write)
```

## Directory Structure

```
searxng-cli/
├── main.go
├── go.mod / go.sum
├── cmd/
│   ├── root.go
│   ├── search.go
│   └── config.go
├── client/
│   └── client.go
├── model/
│   └── result.go
├── formatter/
│   └── formatter.go
├── config/
│   └── config.go
└── SKILL.md              # LLM skill definition for this tool
```

## Data Structures

### Config (`~/.searxng_cli/config.yaml`)

```yaml
default_instance: local
instances:
  local:
    url: http://localhost:8888
```

### Search Result (internal)

```go
type Result struct {
    Title    string `json:"t"`
    URL      string `json:"u"`
    Snippet  string `json:"s"`
    Category string `json:"c"`
    Engine   string `json:"e"`
}
```

## Commands

### `searxng-cli search <query>`

Main search command. Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `compact` | Output format: `compact`, `table`, `urls`, `json` |
| `-c, --categories` | `""` | Comma-separated categories (e.g. `general,news,images`) |
| `--engines` | `""` | Comma-separated engines |
| `--language` | `""` | Language code (e.g. `en`, `ja`) |
| `--time-range` | `""` | `day`, `month`, `year` |
| `-n, --pageno` | `1` | Page number |
| `--instance` | from config | Instance name to use |
| `--max-results` | `0` (all) | Max results to display |

### `searxng-cli config init`

Create `~/.searxng_cli/config.yaml` with default localhost instance.

### `searxng-cli config set-instance <name> <url>`

Add or update an instance entry.

### `searxng-cli config list`

List configured instances.

## Output Formats

### `compact` (default) — JSON Lines, short keys

```jsonl
{"t":"Title","u":"https://example.com","s":"snippet text","c":"general","e":"google"}
{"t":"Another","u":"https://...","s":"more text","c":"news","e":"duckduckgo"}
```

### `table` — human-readable table

```
  # │ Title                     │ URL                              │ Engine
───┼───────────────────────────┼──────────────────────────────────┼────────
  1 │ Title                     │ https://example.com              │ google
  2 │ Another                   │ https://...                      │ duckduckgo
```

### `urls` — URL only, one per line

```
https://example.com
https://...
```

### `json` — raw JSON array (passthrough from SearXNG API)

```json
[{"title":"...", ...}]
```

## SearXNG API Consumption

The client calls `GET /search?q=<query>&format=json` on the configured instance URL, with optional params (`categories`, `engines`, `language`, `time_range`, `pageno`).

SearXNG API returns JSON in this structure:

```json
{
  "results": [
    {"title": "...", "url": "...", "content": "...", "category": "...", "engine": "..."}
  ],
  "answers": [],
  "infoboxes": []
}
```

The client extracts `results` array, maps to internal `Result` struct, and passes to the formatter.

## SKILL.md (for LLM consumption)

Located at project root and also installable to `~/.searxng_cli/SKILL.md`:

```markdown
# SearXNG CLI Skill

## 概要
SearXNGの検索インスタンスに対してWeb検索を実行するCLIツール。
コンパクトなJSON Lines形式で結果を返す。

## 使い方
```
searxng-cli search "<query>"
searxng-cli search "<query>" -c general,news
```

## 出力形式（デフォルト: compact JSONL）
各行が1件の検索結果。
```jsonl
{"t":"Title","u":"https://...","s":"snippet","c":"category","e":"engine"}
```

## 推奨フラグ
- `-c general,news` - カテゴリ指定（結果の品質向上）
- `--max-results 5` - トークン節約
- `--time-range` - 最新情報取得時
```

## Error Handling

1. **Config not found**: On first run, show friendly error: "Config not found. Run `searxng-cli config init` to create one."
2. **API connection failure**: Exit with non-zero code, print error to stderr, no JSON on stdout.
3. **Non-200 response**: Print status code + body snippet to stderr.
4. **No results**: Exit with code 0, empty output (valid JSONL with 0 lines).

## Non-Goals

- REPL/interactive mode (future enhancement)
- Autocomplete support
- Running SearXNG itself (client only)
- Multiple output format streaming
