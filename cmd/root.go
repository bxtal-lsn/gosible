package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ansiblecli",
	Short: "A CLI tool for dynamically running Ansible playbooks",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Add subcommands here
func init() {
	rootCmd.AddCommand(runCmd)
}
