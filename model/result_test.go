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
