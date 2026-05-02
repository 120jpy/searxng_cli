# SearXNG CLI

A CLI tool to search [SearXNG](https://docs.searxng.org/) instances and output results in multiple formats. Designed for both human reading and LLM consumption.

## Installation

### From GitHub Releases

Download the latest binary from the [releases page](https://github.com/120jpy/searxng_cli/releases).

```bash
# macOS (Apple Silicon)
curl -OL https://github.com/120jpy/searxng_cli/releases/download/v0.3.0/searxng-cli_darwin_arm64
chmod +x searxng-cli_darwin_arm64
mv searxng-cli_darwin_arm64 /usr/local/bin/searxng-cli

# macOS (Intel)
curl -OL https://github.com/120jpy/searxng_cli/releases/download/v0.3.0/searxng-cli_darwin_amd64
chmod +x searxng-cli_darwin_amd64
mv searxng-cli_darwin_amd64 /usr/local/bin/searxng-cli

# Linux (amd64)
curl -OL https://github.com/120jpy/searxng_cli/releases/download/v0.3.0/searxng-cli_linux_amd64
chmod +x searxng-cli_linux_amd64
mv searxng-cli_linux_amd64 /usr/local/bin/searxng-cli
```

### Build from Source

```bash
git clone https://github.com/120jpy/searxng_cli.git
cd searxng_cli
go build -o searxng-cli .
```

## Quick Start

Create a default config pointing to `http://127.0.0.1:8888` (local SearXNG):

```bash
searxng-cli config init
```

Or configure a public SearXNG instance:

```bash
searxng-cli config set-instance public https://searx.example.com
searxng-cli config set-instance myinst https://my-searxng.example.com
searxng-cli config list
```

Search:

```bash
searxng-cli search "what is the weather in Tokyo"
```

## Usage

### Search

```bash
searxng-cli search <query> [flags]
```

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `compact` | Output format: `compact`, `table`, `urls`, `json` |
| `--categories` | `-c` | | Categories (e.g. `general,news`) |
| `--engines` | | | Engines (e.g. `google,wikipedia`) |
| `--language` | | | Language code (e.g. `en`, `ja`) |
| `--time-range` | | | Time range: `day`, `month`, `year` |
| `--pageno` | `-n` | `1` | Page number |
| `--instance` | | (default) | Instance name from config |
| `--max-results` | | `0` (all) | Max results to display |
| `--timeout` | `-t` | `30` | Request timeout in seconds |
| `--fetch` | | `false` | Fetch page content with JS rendering (Chrome required) |
| `--fetch-timeout` | | `10` | Per-page fetch timeout in seconds |
| `--fetch-concurrency` | | `3` | Max parallel page fetches |

### Output Formats

**compact** (default) — JSON Lines with short keys for minimal token usage:
```json
{"t":"Title","u":"https://...","s":"snippet","c":"general","e":"google"}
```
With `--fetch`, each line includes the page content as Markdown:
```json
{"t":"Title","u":"https://...","s":"snippet","c":"general","e":"google","b":"# Page Title\\n\\nFull page content..."}
```

**table** — Human-readable table with aligned columns. Falls back to `compact` when `--fetch` is active.

**urls** — One URL per line, for piping into other tools. Falls back to `compact` when `--fetch` is active.

**json** — Pretty-printed JSON array with full key names. With `--fetch`, each result includes a `body` field.

### Progress Display

Progress messages (`Searching <url> ... N results`) are written to stderr so stdout remains clean for piping or redirecting:

```bash
searxng-cli search "query" --max-results 5 -f urls | xargs -I {} curl -O {}
```

### Example

```bash
# Search with categories and time range, show 5 results in table format
searxng-cli search "latest AI news" -c news --time-range day --max-results 5 -f table

# Search and fetch full page content from all results (JS rendering)
searxng-cli search "NVIDIA DGX Spark" --fetch

# Fetch with custom timeout and concurrency
searxng-cli search "quantum computing" --fetch --fetch-timeout 15 --fetch-concurrency 5

# Search with fetch, output as JSON
searxng-cli search "Rust programming" --fetch -f json
```

## Configuration

Config file location: `~/.searxng_cli/config.yaml`

Override with the `SEARXNG_CLI_CONFIG_DIR` environment variable.

```bash
export SEARXNG_CLI_CONFIG_DIR=/path/to/custom/dir
```

```yaml
default_instance: local
instances:
  local:
    url: http://127.0.0.1:8888
  public:
    url: https://searx.example.com
```

### Commands

| Command | Description |
|---------|-------------|
| `config init` | Create default config |
| `config set-instance <name> <url>` | Add or update an instance |
| `config list` | List configured instances |

## Version

```bash
searxng-cli --version
```
