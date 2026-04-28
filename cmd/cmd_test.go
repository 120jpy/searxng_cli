package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestIsConfigInitCmd(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
		want bool
	}{
		{
			name: "config init command",
			cmd:  configInitCmd,
			want: true,
		},
		{
			name: "config set-instance command",
			cmd:  configSetInstanceCmd,
			want: false,
		},
		{
			name: "config list command",
			cmd:  configListCmd,
			want: false,
		},
		{
			name: "search command",
			cmd:  searchCmd,
			want: false,
		},
		{
			name: "nil parent",
			cmd: &cobra.Command{
				Use: "init",
			},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isConfigInitCmd(tc.cmd); got != tc.want {
				t.Errorf("isConfigInitCmd(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestIsConfigInitCmdWithParent(t *testing.T) {
	configCmd := &cobra.Command{Use: "config"}
	initCmd := &cobra.Command{Use: "init"}
	configCmd.AddCommand(initCmd)

	if !isConfigInitCmd(initCmd) {
		t.Error("expected isConfigInitCmd to be true for config init subcommand")
	}
}

func TestIsConfigInitCmdWrongParent(t *testing.T) {
	otherCmd := &cobra.Command{Use: "other"}
	initCmd := &cobra.Command{Use: "init"}
	otherCmd.AddCommand(initCmd)

	if isConfigInitCmd(initCmd) {
		t.Error("expected isConfigInitCmd to be false for init under non-config parent")
	}
}

func TestConfigSetInstanceInvalidURL(t *testing.T) {
	err := configSetInstanceCmd.RunE(configSetInstanceCmd, []string{"test", "not a url"})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}
