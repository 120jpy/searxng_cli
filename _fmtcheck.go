package main

import (
	"fmt"
	"github.com/tomo/searxng-cli/formatter"
	"github.com/tomo/searxng-cli/model"
)

func main() {
	results := []model.Result{
		{Title: "Title1", URL: "https://a.com", Snippet: "snippet1", Category: "general", Engine: "google"},
		{Title: "Title2", URL: "https://b.com", Snippet: "snippet2", Category: "news", Engine: "duckduckgo"},
	}
	fmt.Println("=== COMPACT ===")
	fmt.Print(formatter.FormatResults(results, formatter.FormatCompact, 0))
	fmt.Println("=== TABLE ===")
	fmt.Print(formatter.FormatResults(results, formatter.FormatTable, 0))
	fmt.Println("=== URLS ===")
	fmt.Print(formatter.FormatResults(results, formatter.FormatURLs, 0))
	fmt.Println("=== JSON ===")
	fmt.Print(formatter.FormatResults(results, formatter.FormatJSON, 0))
	fmt.Println("=== TRUNCATED (1) ===")
	fmt.Print(formatter.FormatResults(results, formatter.FormatCompact, 1))
}
