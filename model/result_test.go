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

func TestResultBodyOmitEmpty(t *testing.T) {
	r := Result{
		Title: "Test", URL: "https://example.com",
		Snippet: "snip", Category: "general", Engine: "google",
	}
	data, _ := json.Marshal(r)
	if _, ok := jsonGet(string(data), "b"); ok {
		t.Fatal("Body should be omitted when empty (omitempty), got key 'b'")
	}
}

func TestResultBodyPresent(t *testing.T) {
	r := Result{
		Title: "Test", URL: "https://example.com",
		Snippet: "snip", Category: "general", Engine: "google",
		Body: "markdown content",
	}
	data, _ := json.Marshal(r)
	val, ok := jsonGet(string(data), "b")
	if !ok {
		t.Fatal("Body should be present as key 'b' when non-empty")
	}
	if val != "markdown content" {
		t.Fatalf("Body value = %q, want %q", val, "markdown content")
	}
}

// jsonGet parses a JSON object line and returns the value for a key as string.
func jsonGet(jsonStr, key string) (string, bool) {
	var m map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return "", false
	}
	v, ok := m[key]
	return v, ok
}
