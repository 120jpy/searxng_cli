# SearXNG CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a single-binary Go CLI tool that queries SearXNG instances and outputs results in LLM-friendly compact JSON Lines format.

**Architecture:** CLI entry point via cobra, HTTP requests via native `net/http`, YAML config via `gopkg.in/yaml.v3`. The flow is: `cmd/search.go` → `client/client.go` (HTTP) → `formatter/formatter.go` (output).

**Tech Stack:** Go 1.24, cobra (CLI), yaml.v3 (config), standard library (HTTP).

---

### Task 1: Project scaffold and module init

**Files:**
- Create: `go.mod`
- Create: `main.go`

- [ ] **Step 1: Initialize Go module and install dependencies**

Run:
```bash
cd /Users/tomo/Documents/Git/searxng_cli
go mod init github.com/tomo/searxng-cli
go get github.com/spf13/cobra
go get gopkg.in/yaml.v3
```
Expected: `go.mod` and `go.sum` created with cobra and yaml.v3 dependencies.

- [ ] **Step 2: Create main.go entry point**

```go
package main

import "github.com/tomo/searxng-cli/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 3: Create cmd/root.go with cobra root command**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tomo/searxng-cli/config"
)

var (
	cfg         *config.Config
	cfgFileFlag string
)

var rootCmd = &cobra.Command{
	Use:   "searxng-cli",
	Short: "SearXNG CLI - Web search via SearXNG instances",
	Long: `A CLI tool to query SearXNG instances and output results
in LLM-friendly formats.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("config load failed: %w\nRun 'searxng-cli config init' to create one", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFileFlag, "config", "", "config file path (default ~/.searxng_cli/config.yaml)")
}
```

- [ ] **Step 4: Verify it compiles**

Run:
```bash
go build ./...
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum main.go cmd/root.go
git commit -m "feat: initial project scaffold with cobra"
```

---

### Task 2: Config package

**Files:**
- Create: `config/config.go`
- Test: `config/config_test.go`

- [ ] **Step 1: Write the config package**

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Instance struct {
	URL string `yaml:"url"`
}

type Config struct {
	DefaultInstance string              `yaml:"default_instance"`
	Instances       map[string]Instance `yaml:"instances"`
}

func configDir() string {
	if v := os.Getenv("SEARXNG_CLI_CONFIG_DIR"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".searxng_cli")
}

func configPath() string {
	return filepath.Join(configDir(), "config.yaml")
}

func Load() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	if cfg.Instances == nil {
		cfg.Instances = make(map[string]Instance)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

func (c *Config) GetInstance(name string) (*Instance, error) {
	if name == "" {
		name = c.DefaultInstance
	}
	inst, ok := c.Instances[name]
	if !ok {
		return nil, fmt.Errorf("instance %q not found in config", name)
	}
	return &inst, nil
}

// DefaultConfig returns a config with a default localhost instance.
func DefaultConfig() *Config {
	return &Config{
		DefaultInstance: "local",
		Instances: map[string]Instance{
			"local": {URL: "http://127.0.0.1:8888"},
		},
	}
}
```

- [ ] **Step 2: Write tests**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirEnvVar(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEARXNG_CLI_CONFIG_DIR", dir)
	got := configDir()
	if got != dir {
		t.Fatalf("configDir() = %s, want %s", got, dir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEARXNG_CLI_CONFIG_DIR", dir)

	cfg := DefaultConfig()
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DefaultInstance != "local" {
		t.Fatalf("DefaultInstance = %q, want local", loaded.DefaultInstance)
	}
	if loaded.Instances["local"].URL != "http://127.0.0.1:8888" {
		t.Fatalf("URL = %q, want http://127.0.0.1:8888", loaded.Instances["local"].URL)
	}
}

func TestGetInstanceDefault(t *testing.T) {
	cfg := DefaultConfig()
	inst, err := cfg.GetInstance("")
	if err != nil {
		t.Fatal(err)
	}
	if inst.URL != "http://127.0.0.1:8888" {
		t.Fatalf("URL = %q", inst.URL)
	}
}

func TestGetInstanceNotFound(t *testing.T) {
	cfg := DefaultConfig()
	_, err := cfg.GetInstance("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SEARXNG_CLI_CONFIG_DIR", dir)
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when config missing")
	}
}
```

- [ ] **Step 3: Run tests**

Run:
```bash
go test ./config/ -v
```
Expected: all 5 tests pass.

- [ ] **Step 4: Commit**

```bash
git add config/config.go config/config_test.go
git commit -m "feat: config package with YAML read/write"
```

---

### Task 3: Model package

**Files:**
- Create: `model/result.go`
- Test: `model/result_test.go`

- [ ] **Step 1: Create model package**

```go
package model

type Result struct {
	Title    string `json:"t"`
	URL      string `json:"u"`
	Snippet  string `json:"s"`
	Category string `json:"c"`
	Engine   string `json:"e"`
}
```

- [ ] **Step 2: Write tests**

```go
package model

import (
	"encoding/json"
	"testing"
)

func TestResultJSONTags(t *testing.T) {
	r := Result{
		Title:    "Hello World",
		URL:      "https://example.com",
		Snippet:  "A snippet",
		Category: "general",
		Engine:   "google",
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	json.Unmarshal(data, &got)
	if got["t"] != "Hello World" {
		t.Fatalf("key 't' = %q, want 'Hello World'", got["t"])
	}
	if got["u"] != "https://example.com" {
		t.Fatalf("key 'u' = %q", got["u"])
	}
	if _, ok := got["title"]; ok {
		t.Fatal("should NOT have key 'title', only short key 't'")
	}
}
```

- [ ] **Step 3: Run tests**

Run:
```bash
go test ./model/ -v
```
Expected: pass.

- [ ] **Step 4: Commit**

```bash
git add model/result.go model/result_test.go
git commit -m "feat: model package with short JSON tags"
```

---

### Task 4: Client package (HTTP API call)

**Files:**
- Create: `client/client.go`
- Test: `client/client_test.go`

- [ ] **Step 1: Create client package**

```go
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tomo/searxng-cli/model"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type SearchParams struct {
	Query      string
	Categories string
	Engines    string
	Language   string
	TimeRange  string
	Pageno     int
}

// searxngAPIResponse mirrors the relevant parts of the SearXNG JSON API response.
type searxngAPIResponse struct {
	Results []struct {
		Title    string `json:"title"`
		URL      string `json:"url"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Engine   string `json:"engine"`
	} `json:"results"`
}

func (c *Client) Search(params SearchParams) ([]model.Result, error) {
	u, err := url.Parse(c.BaseURL + "/search")
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	q := u.Query()
	q.Set("q", params.Query)
	q.Set("format", "json")
	if params.Categories != "" {
		q.Set("categories", params.Categories)
	}
	if params.Engines != "" {
		q.Set("engines", params.Engines)
	}
	if params.Language != "" {
		q.Set("language", params.Language)
	}
	if params.TimeRange != "" {
		q.Set("time_range", params.TimeRange)
	}
	if params.Pageno > 0 {
		q.Set("pageno", fmt.Sprintf("%d", params.Pageno))
	}
	u.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var apiResp searxngAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	results := make([]model.Result, 0, len(apiResp.Results))
	for _, r := range apiResp.Results {
		results = append(results, model.Result{
			Title:    r.Title,
			URL:      r.URL,
			Snippet:  r.Content,
			Category: r.Category,
			Engine:   r.Engine,
		})
	}
	return results, nil
}
```

- [ ] **Step 2: Write test (using httptest)**

```go
package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tomo/searxng-cli/model"
)

func TestSearchSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "json" {
			t.Fatalf("expected format=json, got %q", r.URL.Query().Get("format"))
		}
		if r.URL.Query().Get("q") != "test query" {
			t.Fatalf("expected q=test query, got %q", r.URL.Query().Get("q"))
		}
		resp := map[string]interface{}{
			"results": []map[string]string{
				{"title": "Title1", "url": "https://a.com", "content": "snippet1", "category": "general", "engine": "google"},
				{"title": "Title2", "url": "https://b.com", "content": "snippet2", "category": "news", "engine": "duckduckgo"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := New(ts.URL)
	results, err := c.Search(SearchParams{Query: "test query"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	expected := []model.Result{
		{Title: "Title1", URL: "https://a.com", Snippet: "snippet1", Category: "general", Engine: "google"},
		{Title: "Title2", URL: "https://b.com", Snippet: "snippet2", Category: "news", Engine: "duckduckgo"},
	}
	for i, e := range expected {
		if results[i] != e {
			t.Fatalf("result[%d] = %+v, want %+v", i, results[i], e)
		}
	}
}

func TestSearchServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	c := New(ts.URL)
	_, err := c.Search(SearchParams{Query: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchBadURL(t *testing.T) {
	c := New("http://[::1]:nonexistent")
	_, err := c.Search(SearchParams{Query: "test"})
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestSearchCategoriesParam(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("categories") != "general,news" {
			t.Fatalf("categories = %q", r.URL.Query().Get("categories"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(searxngAPIResponse{})
	}))
	defer ts.Close()

	c := New(ts.URL)
	_, err := c.Search(SearchParams{Query: "test", Categories: "general,news"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchEnginesParam(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("engines") != "google,wikipedia" {
			t.Fatalf("engines = %q", r.URL.Query().Get("engines"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(searxngAPIResponse{})
	}))
	defer ts.Close()

	c := New(ts.URL)
	_, err := c.Search(SearchParams{Query: "test", Engines: "google,wikipedia"})
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 3: Run tests**

Run:
```bash
go test ./client/ -v
```
Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add client/client.go client/client_test.go
git commit -m "feat: HTTP client for SearXNG search API"
```

---

### Task 5: Formatter package

**Files:**
- Create: `formatter/formatter.go`
- Test: `formatter/formatter_test.go`

- [ ] **Step 1: Create formatter package**

```go
package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/tomo/searxng-cli/model"
)

type Format string

const (
	FormatCompact Format = "compact"
	FormatTable   Format = "table"
	FormatURLs    Format = "urls"
	FormatJSON    Format = "json"
)

var ValidFormats = []string{string(FormatCompact), string(FormatTable), string(FormatURLs), string(FormatJSON)}

func FormatResults(results []model.Result, format Format, maxResults int) string {
	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}

	switch format {
	case FormatCompact:
		return formatCompact(results)
	case FormatTable:
		return formatTable(results)
	case FormatURLs:
		return formatURLs(results)
	case FormatJSON:
		return formatJSON(results)
	default:
		return formatCompact(results)
	}
}

func formatCompact(results []model.Result) string {
	var b strings.Builder
	for _, r := range results {
		data, _ := json.Marshal(r)
		b.Write(data)
		b.WriteByte('\n')
	}
	return b.String()
}

func formatTable(results []model.Result) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tTitle\tURL\tEngine")
	fmt.Fprintln(w, "-\t-----\t---\t------")
	for i, r := range results {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", i+1, truncate(r.Title, 40), truncate(r.URL, 50), r.Engine)
	}
	w.Flush()
	return b.String()
}

func formatURLs(results []model.Result) string {
	var b strings.Builder
	for _, r := range results {
		b.WriteString(r.URL)
		b.WriteByte('\n')
	}
	return b.String()
}

func formatJSON(results []model.Result) string {
	// raw JSON array, preserving SearXNG's field names
	type rawResult struct {
		Title    string `json:"title"`
		URL      string `json:"url"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Engine   string `json:"engine"`
	}
	raw := make([]rawResult, len(results))
	for i, r := range results {
		raw[i] = rawResult{
			Title:    r.Title,
			URL:      r.URL,
			Content:  r.Snippet,
			Category: r.Category,
			Engine:   r.Engine,
		}
	}
	data, _ := json.MarshalIndent(raw, "", "  ")
	return string(data)
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-1] + "…"
	}
	return s
}
```

- [ ] **Step 2: Write tests**

```go
package formatter

import (
	"strings"
	"testing"

	"github.com/tomo/searxng-cli/model"
)

func makeSampleResults() []model.Result {
	return []model.Result{
		{Title: "Title1", URL: "https://a.com", Snippet: "snippet1", Category: "general", Engine: "google"},
		{Title: "Title2", URL: "https://b.com", Snippet: "snippet2", Category: "news", Engine: "duckduckgo"},
	}
}

func TestCompactFormat(t *testing.T) {
	results := makeSampleResults()
	out := FormatResults(results, FormatCompact, 0)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	// each line should be valid JSON with short keys
	for _, line := range lines {
		if !strings.Contains(line, `"t":"`) {
			t.Fatalf("line missing short key 't': %s", line)
		}
		if !strings.Contains(line, `"u":"`) {
			t.Fatalf("line missing short key 'u': %s", line)
		}
	}
}

func TestTableFormat(t *testing.T) {
	results := makeSampleResults()
	out := FormatResults(results, FormatTable, 0)
	if !strings.Contains(out, "Title1") {
		t.Fatalf("table missing Title1: %s", out)
	}
	if !strings.Contains(out, "https://a.com") {
		t.Fatalf("table missing URL")
	}
}

func TestURLsFormat(t *testing.T) {
	results := makeSampleResults()
	out := FormatResults(results, FormatURLs, 0)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0] != "https://a.com" {
		t.Fatalf("first URL = %q", lines[0])
	}
	if lines[1] != "https://b.com" {
		t.Fatalf("second URL = %q", lines[1])
	}
}

func TestJSONFormat(t *testing.T) {
	results := makeSampleResults()
	out := FormatResults(results, FormatJSON, 0)
	if !strings.Contains(out, `"title"`) {
		t.Fatalf("json format should use long keys: %s", out)
	}
	if !strings.Contains(out, `"content"`) {
		t.Fatalf("json format missing content: %s", out)
	}
}

func TestMaxResults(t *testing.T) {
	results := makeSampleResults()
	out := FormatResults(results, FormatCompact, 1)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines with max=1, want 1", len(lines))
	}
}
```

- [ ] **Step 3: Run tests**

Run:
```bash
go test ./formatter/ -v
```
Expected: all tests pass.

- [ ] **Step 4: Commit**

```bash
git add formatter/formatter.go formatter/formatter_test.go
git commit -m "feat: formatter with compact/table/urls/json output"
```

---

### Task 6: Config commands (config init / set-instance / list)

**Files:**
- Create: `cmd/config.go`

- [ ] **Step 1: Create config subcommands**

```go
package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tomo/searxng-cli/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default config at ~/.searxng_cli/config.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Default config created at ~/.searxng_cli/config.yaml")
		return nil
	},
}

var configSetInstanceCmd = &cobra.Command{
	Use:   "set-instance <name> <url>",
	Short: "Add or update an instance",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, url := args[0], args[1]
		cfg.Instances[name] = config.Instance{URL: url}
		if len(cfg.Instances) == 1 {
			cfg.DefaultInstance = name
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Instance %q set to %s\n", name, url)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(cfg.Instances) == 0 {
			fmt.Fprintln(os.Stderr, "No instances configured.")
			return nil
		}
		names := make([]string, 0, len(cfg.Instances))
		for n := range cfg.Instances {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			mark := " "
			if n == cfg.DefaultInstance {
				mark = "*"
			}
			fmt.Printf("%s %s: %s\n", mark, n, cfg.Instances[n].URL)
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetInstanceCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}
```

- [ ] **Step 2: Verify compilation**

Run:
```bash
go build ./...
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/config.go
git commit -m "feat: config init/set-instance/list commands"
```

---

### Task 7: Search command

**Files:**
- Create: `cmd/search.go`

- [ ] **Step 1: Create search subcommand**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tomo/searxng-cli/client"
	"github.com/tomo/searxng-cli/formatter"
)

var searchFlags struct {
	format     string
	categories string
	engines    string
	language   string
	timeRange  string
	pageno     int
	instance   string
	maxResults int
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search via SearXNG instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// resolve instance
		inst, err := cfg.GetInstance(searchFlags.instance)
		if err != nil {
			return err
		}

		// validate format
		valid := false
		for _, f := range formatter.ValidFormats {
			if searchFlags.format == f {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid format %q, valid: %v", searchFlags.format, formatter.ValidFormats)
		}

		// search
		c := client.New(inst.URL)
		results, err := c.Search(client.SearchParams{
			Query:      args[0],
			Categories: searchFlags.categories,
			Engines:    searchFlags.engines,
			Language:   searchFlags.language,
			TimeRange:  searchFlags.timeRange,
			Pageno:     searchFlags.pageno,
		})
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		// format and output
		out := formatter.FormatResults(results, formatter.Format(searchFlags.format), searchFlags.maxResults)
		fmt.Print(out)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchFlags.format, "format", "f", "compact", "Output format: compact, table, urls, json")
	searchCmd.Flags().StringVarP(&searchFlags.categories, "categories", "c", "", "Comma-separated categories (e.g. general,news)")
	searchCmd.Flags().StringVar(&searchFlags.engines, "engines", "", "Comma-separated engines (e.g. google,wikipedia)")
	searchCmd.Flags().StringVar(&searchFlags.language, "language", "", "Language code (e.g. en, ja)")
	searchCmd.Flags().StringVar(&searchFlags.timeRange, "time-range", "", "Time range: day, month, year")
	searchCmd.Flags().IntVarP(&searchFlags.pageno, "pageno", "n", 1, "Page number")
	searchCmd.Flags().StringVar(&searchFlags.instance, "instance", "", "Instance name (default from config)")
	searchCmd.Flags().IntVar(&searchFlags.maxResults, "max-results", 0, "Max results to display (0 = all)")
	rootCmd.AddCommand(searchCmd)
}
```

- [ ] **Step 2: Verify compilation**

Run:
```bash
go build ./...
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/search.go
git commit -m "feat: search command with format/categories/engine flags"
```

---

### Task 8: Full build test

**Files:** (no new files)

- [ ] **Step 1: Build the binary**

Run:
```bash
go build -o searxng-cli .
```
Expected: binary `searxng-cli` created.

- [ ] **Step 2: Run all tests**

Run:
```bash
go test ./... -v
```
Expected: all tests pass.

- [ ] **Step 3: Verify config init works**

Run:
```bash
SEARXNG_CLI_CONFIG_DIR=/tmp/test_searxng_cli ./searxng-cli config init
```
Expected: "Default config created" message.

- [ ] **Step 4: Verify config list works**

```bash
SEARXNG_CLI_CONFIG_DIR=/tmp/test_searxng_cli ./searxng-cli config list
```
Expected: `* local: http://127.0.0.1:8888`

- [ ] **Step 5: Verify config set-instance works**

```bash
SEARXNG_CLI_CONFIG_DIR=/tmp/test_searxng_cli ./searxng-cli config set-instance myinst https://searx.example.com
SEARXNG_CLI_CONFIG_DIR=/tmp/test_searxng_cli ./searxng-cli config list
```
Expected: both instances listed, `myinst` shown.

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "chore: final build and test verification"
```

---

### Task 9: SKILL.md

**Files:**
- Create: `SKILL.md`

- [ ] **Step 1: Create SKILL.md**

```markdown
# SearXNG CLI Skill

## 概要
SearXNG の検索インスタンスに対して Web 検索を実行する CLI ツール。
コンパクトな JSON Lines 形式で結果を返す。

## 使い方

```bash
# 検索（デフォルト出力: compact JSONL）
searxng-cli search "<query>"

# カテゴリ指定
searxng-cli search "<query>" -c general,news

# エンジン指定
searxng-cli search "<query>" --engines google,wikipedia

# 出力フォーマット変更
searxng-cli search "<query>" -f table
searxng-cli search "<query>" -f urls

# 件数制限
searxng-cli search "<query>" --max-results 5

# 時間範囲
searxng-cli search "<query>" --time-range day

# インスタンス切り替え
searxng-cli search "<query>" --instance myinst
```

## 出力形式（デフォルト: compact JSONL）

各行が 1 件の検索結果。キーは短縮（`t`=title, `u`=url, `s`=snippet, `c`=category, `e`=engine）。

```jsonl
{"t":"Title","u":"https://...","s":"snippet","c":"general","e":"google"}
```

## 推奨フラグ

- `-c general,news` - カテゴリ指定（結果の品質向上）
- `--max-results 5` - トークン節約
- `--time-range day` - 最新情報取得時

## 設定

初回実行前に設定が必要:

```bash
searxng-cli config init                       # デフォルト設定作成 (localhost:8888)
searxng-cli config set-instance public https://searx.example.com  # 公開インスタンス追加
searxng-cli config list                       # 設定一覧
```

環境変数 `SEARXNG_CLI_CONFIG_DIR` で設定ディレクトリを変更可能。
```

- [ ] **Step 2: Commit**

```bash
git add SKILL.md
git commit -m "docs: add SKILL.md for LLM consumption"
```
