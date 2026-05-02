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

func TestCompactFormatWithBody(t *testing.T) {
	results := []model.Result{
		{Title: "T1", URL: "https://a.com", Snippet: "s1", Category: "general", Engine: "google", Body: "Hello\nWorld"},
		{Title: "T2", URL: "https://b.com", Snippet: "s2", Category: "news", Engine: "duckduckgo"},
	}
	out := FormatResults(results, FormatCompact, 0)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if !strings.Contains(lines[0], "\"b\":\"Hello") {
		t.Fatalf("first line missing body key 'b': %s", lines[0])
	}
	if strings.Contains(lines[1], "\"b\"") {
		t.Fatalf("second line should NOT have body key (omitempty), got: %s", lines[1])
	}
}

func TestJSONFormatWithBody(t *testing.T) {
	results := []model.Result{
		{Title: "T1", URL: "https://a.com", Snippet: "s1", Category: "general", Engine: "google", Body: "Hello World"},
	}
	out := FormatResults(results, FormatJSON, 0)
	if !strings.Contains(out, "\"body\": \"Hello World\"") && !strings.Contains(out, "\"body\": \"Hello") {
		t.Fatalf("json should include body key: %s", out)
	}
}

func TestTableAutoSwitchToCompactWithBody(t *testing.T) {
	results := []model.Result{
		{Title: "T1", URL: "https://a.com", Snippet: "s1", Category: "general", Engine: "google", Body: "content"},
	}
	out := FormatResults(results, FormatTable, 0)
	if !strings.Contains(out, "\"t\":\"") || !strings.Contains(out, "\"b\":\"content\"") {
		t.Fatalf("table with body should auto-switch to compact: %s", out)
	}
}

func TestURLsAutoSwitchToCompactWithBody(t *testing.T) {
	results := []model.Result{
		{Title: "T1", URL: "https://a.com", Snippet: "s1", Category: "general", Engine: "google", Body: "content"},
	}
	out := FormatResults(results, FormatURLs, 0)
	if !strings.Contains(out, "\"t\":\"") || !strings.Contains(out, "\"b\":\"content\"") {
		t.Fatalf("urls with body should auto-switch to compact: %s", out)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		wantLen int
		want    string
	}{
		{"short", 10, 5, "short"},
		{"this is a long string", 10, 10, "this is a…"},
		{"", 5, 0, ""},
		{"こんにちは世界", 10, 7, "こんにちは世界"},
		{"こんにちは世界", 5, 5, "こんにち…"},
	}
	for _, tc := range tests {
		got := truncate(tc.input, tc.maxLen)
		if len([]rune(got)) != tc.wantLen {
			t.Fatalf("truncate(%q, %d) = %q (rune len %d), want rune len %d", tc.input, tc.maxLen, got, len([]rune(got)), tc.wantLen)
		}
		if got != tc.want {
			t.Fatalf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}
