package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tomo/searxng-cli/config"
)

var (
	cfg         *config.Config
	cfgFileFlag string
)

var rootCmd = &cobra.Command{
	Use:   "searxng-cli",
	Short: "SearXNG CLI - Web search via SearXNG instances",
	Long: `A CLI tool to query SearXNG instances and output results
in LLM-friendly formats.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if cfgFileFlag != "" {
			cfg, err = config.LoadFrom(cfgFileFlag)
		} else {
			cfg, err = config.Load()
		}
		if err != nil {
			return fmt.Errorf("config load failed: %w\nRun 'searxng-cli config init' to create one", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFileFlag, "config", "", "config file path (default ~/.searxng_cli/config.yaml)")
}
