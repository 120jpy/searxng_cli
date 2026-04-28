package cmd

import (
	"fmt"

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
