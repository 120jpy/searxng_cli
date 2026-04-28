package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tomo/searxng-cli/config"
)

const versionTemplate = "searxng-cli {{.Version}} (commit: {{.Commit}}, built: {{.Date}})\n"

var (
	cfg         *config.Config
	cfgFileFlag string
	Version     = "dev"
	Commit      = "none"
	Date        = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "searxng-cli",
	Short:   "SearXNG CLI - Web search via SearXNG instances",
	Long:    `A CLI tool to query SearXNG instances and output results in LLM-friendly formats.`,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for version and config init
		if isConfigInitCmd(cmd) || cmd.Name() == "version" {
			return nil
		}
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

func isConfigInitCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "init" && cmd.Parent() != nil && cmd.Parent().Name() == "config"
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFileFlag, "config", "", "config file path (default ~/.searxng_cli/config.yaml)")
	rootCmd.SetVersionTemplate(fmt.Sprintf("searxng-cli {{.Version}} (commit: %s, built: %s)\n", Commit, Date))
}
