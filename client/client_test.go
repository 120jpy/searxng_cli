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
		if r.Header.Get("User-Agent") != "searxng-cli/1.0" {
			t.Fatalf("expected User-Agent searxng-cli/1.0, got %q", r.Header.Get("User-Agent"))
		}
		resp := searxngAPIResponse{
			Results: []struct {
				Title    string `json:"title"`
				URL      string `json:"url"`
				Content  string `json:"content"`
				Category string `json:"category"`
				Engine   string `json:"engine"`
			}{
				{Title: "Title1", URL: "https://a.com", Content: "snippet1", Category: "general", Engine: "google"},
				{Title: "Title2", URL: "https://b.com", Content: "snippet2", Category: "news", Engine: "duckduckgo"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := New(ts.URL, 10)
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

	c := New(ts.URL, 10)
	_, err := c.Search(SearchParams{Query: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchBadURL(t *testing.T) {
	c := New("http://[::1]:nonexistent", 10)
	_, err := c.Search(SearchParams{Query: "test"})
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestSearchTrailingSlash(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("expected /search, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(searxngAPIResponse{})
	}))
	defer ts.Close()

	c := New(ts.URL+"/", 10)
	_, err := c.Search(SearchParams{Query: "test"})
	if err != nil {
		t.Fatal(err)
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

	c := New(ts.URL, 10)
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

	c := New(ts.URL, 10)
	_, err := c.Search(SearchParams{Query: "test", Engines: "google,wikipedia"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchRedirectFollowed(t *testing.T) {
	var seenFinal bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/search?q=test&format=json", http.StatusMovedPermanently)
			return
		}
		seenFinal = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(searxngAPIResponse{})
	}))
	defer ts.Close()

	c := New(ts.URL+"/redirect", 10)
	_, err := c.Search(SearchParams{Query: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !seenFinal {
		t.Fatal("redirect was not followed")
	}
}
