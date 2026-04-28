package cmd

import (
	"fmt"
	"net/url"
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
		cfg = config.DefaultConfig()
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
		name, rawURL := args[0], args[1]
		if _, err := url.ParseRequestURI(rawURL); err != nil {
			return fmt.Errorf("invalid URL %q: %w", rawURL, err)
		}
		cfg.Instances[name] = config.Instance{URL: rawURL}
		if len(cfg.Instances) == 1 {
			cfg.DefaultInstance = name
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Instance %q set to %s\n", name, rawURL)
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
