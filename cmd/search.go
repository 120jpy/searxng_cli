package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomo/searxng-cli/client"
	"github.com/tomo/searxng-cli/fetcher"
	"github.com/tomo/searxng-cli/formatter"
)

var searchFlags struct {
	format           string
	categories       string
	engines          string
	language         string
	timeRange        string
	pageno           int
	instance         string
	maxResults       int
	timeout          int
	fetch            bool
	fetchTimeout     int
	fetchConcurrency int
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search via SearXNG instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inst, err := cfg.GetInstance(searchFlags.instance)
		if err != nil {
			return err
		}

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

		fmt.Fprintf(cmd.ErrOrStderr(), "Searching %s ... ", inst.URL)

		c := client.New(inst.URL, searchFlags.timeout)
		results, err := c.Search(client.SearchParams{
			Query:      args[0],
			Categories: searchFlags.categories,
			Engines:    searchFlags.engines,
			Language:   searchFlags.language,
			TimeRange:  searchFlags.timeRange,
			Pageno:     searchFlags.pageno,
		})
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "error")
			return fmt.Errorf("search failed: %w", err)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "%d results\n", len(results))

		if searchFlags.fetch && len(results) > 0 {
			fmt.Fprintf(cmd.ErrOrStderr(), "Fetching %d pages ... ", len(results))
			urls := make([]string, len(results))
			for i, r := range results {
				urls[i] = r.URL
			}
			bodies := fetcher.FetchURLs(urls, searchFlags.fetchTimeout, searchFlags.fetchConcurrency)
			for i, r := range results {
				if body, ok := bodies[r.URL]; ok {
					results[i].Body = body
				}
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "done\n")
		}

		fmt.Print(formatter.FormatResults(results, formatter.Format(searchFlags.format), searchFlags.maxResults))
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
	searchCmd.Flags().IntVarP(&searchFlags.timeout, "timeout", "t", 30, "Request timeout in seconds")
	searchCmd.Flags().BoolVar(&searchFlags.fetch, "fetch", false, "Fetch page content with JS rendering")
	searchCmd.Flags().IntVar(&searchFlags.fetchTimeout, "fetch-timeout", 10, "Per-page fetch timeout in seconds")
	searchCmd.Flags().IntVar(&searchFlags.fetchConcurrency, "fetch-concurrency", 3, "Max parallel page fetches")
	rootCmd.AddCommand(searchCmd)
}
