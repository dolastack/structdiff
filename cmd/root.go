package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/dolastack/structdiff/compare"
	"github.com/dolastack/structdiff/output"
	"github.com/spf13/cobra"
)

var (
	format       string
	outputFormat string
	color        bool
	ignoreCase   bool
	skipValidate bool
	timeout      time.Duration
	maxSize      int64
	username     string
	password     string
	token        string
)

var rootCmd = &cobra.Command{
	Use:   "structdiff <file1> <file2>",
	Short: "Compare structured configuration files",
	Long: `StructDiff compares configuration files in various formats and shows
differences with automatic validation and helpful error messages.`,
	Args: cobra.ExactArgs(2),
	Run:  runComparison,
}

func runComparison(cmd *cobra.Command, args []string) {
	file1, file2 := args[0], args[1]

	// Check if both files are stdin
	if file1 == "-" && file2 == "-" {
		fmt.Fprintln(os.Stderr, "Error: cannot read both files from stdin")
		os.Exit(1)
	}

	config := compare.RemoteConfig{
		Timeout:      timeout,
		MaxFileSize:  maxSize,
		Username:     username,
		Password:     password,
		Token:        token,
		SkipValidate: skipValidate,
	}

	if format == "auto" {
		detected, err := compare.DetectFormat(file1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting format: %v\n", err)
			os.Exit(1)
		}
		format = detected
	}

	diff, err := compare.CompareFiles(file1, file2, format, ignoreCase, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error comparing files: %v\n", err)
		os.Exit(1)
	}

	var result string
	switch outputFormat {
	case "json":
		result, err = output.FormatJSON(diff)
	default:
		result, err = output.FormatText(diff, color)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&format, "format", "f", "auto", "Input format (json|yaml|toml|xml|ini|csv|hcl|auto)")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json)")
	rootCmd.Flags().BoolVar(&color, "color", true, "Enable colored output")
	rootCmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Ignore case differences")
	rootCmd.Flags().BoolVar(&skipValidate, "skip-validate", false, "Skip file validation")
	rootCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "Request timeout")
	rootCmd.Flags().Int64Var(&maxSize, "max-size", 10*1024*1024, "Max file size in bytes")
	rootCmd.Flags().StringVar(&username, "username", "", "Basic auth username")
	rootCmd.Flags().StringVar(&password, "password", "", "Basic auth password")
	rootCmd.Flags().StringVar(&token, "token", "", "Bearer token")
}
