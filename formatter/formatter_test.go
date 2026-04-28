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
