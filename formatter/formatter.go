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
	runes := []rune(s)
	if len(runes) > maxLen {
		if maxLen <= 3 {
			return string(runes[:maxLen])
		}
		return string(runes[:maxLen-1]) + "…"
	}
	return s
}
