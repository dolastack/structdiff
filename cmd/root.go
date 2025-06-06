// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "structdiff",
	Short: "A CLI tool to compare structured data files like JSON, YAML, TOML, XML, INI, CSV",
	Long: `structdiff compares two structured files and shows differences.
Supports:
  - JSON
  - YAML
  - TOML
  - XML
  - INI
  - CSV`,
}

func Execute() error {
	return rootCmd.Execute()
}
