package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dolastack/structdiff/internal/parser"
	"github.com/dolastack/structdiff/pkg/diff"
	"github.com/spf13/cobra"
)

var (
	checkFlag     bool
	quietFlag     bool
	filterPaths   []string
	basicUser     string
	basicPass     string
	bearerToken   string
	customHeaders []string
)

var compareCmd = &cobra.Command{
	Use:   "compare [file1] [file2]",
	Short: "Compare two structured data files",
	Long: `Compare two structured files like JSON, YAML, TOML, XML, INI, or CSV.
Supports local files, stdin (-), and remote URLs.`,
	Example: `structdiff compare file1.yaml file2.yaml --filter=user.name
structdiff compare https://example.com/file1.json  https://example.com/file2.json  --basic-username=admin --basic-password=secret`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		file1 := args[0]
		file2 := args[1]

		headersMap := make(map[string]string)
		for _, h := range customHeaders {
			parts := strings.SplitN(h, "=", 2)
			if len(parts) == 2 {
				headersMap[parts[0]] = resolveValue(parts[1])
			} else if len(parts) == 1 {
				headersMap[parts[0]] = ""
			}
		}

		opts := parser.ParseOptions{
			BasicUser:     basicUser,
			BasicPassword: resolveValue(basicPass),
			BearerToken:   resolveValue(bearerToken),
			CustomHeaders: headersMap,
		}

		d1, err := parser.ParseFile(file1, opts)
		if err != nil {
			return fmt.Errorf("error parsing file1: %v", err)
		}

		d2, err := parser.ParseFile(file2, opts)
		if err != nil {
			return fmt.Errorf("error parsing file2: %v", err)
		}

		diffOutput := diff.Compare(d1, d2, filterPaths)

		if diffOutput == "" {
			if quietFlag {
				return nil
			}
			fmt.Fprintln(os.Stdout, "No differences found.")
			return nil
		}

		if quietFlag {
			return nil
		}

		count := countLines(diffOutput)
		fmt.Fprintf(os.Stdout, "Found %d differences\n", count)
		fmt.Fprint(os.Stdout, diffOutput)

		if checkFlag {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)

	// Diff Options
	compareCmd.Flags().BoolVarP(&checkFlag, "check", "c", false,
		"Exit with non-zero code if differences found")
	compareCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false,
		"Only output summary (use with --check)")
	compareCmd.Flags().StringArrayVarP(&filterPaths, "filter", "f", nil,
		"Only show diffs under this key path (e.g., user.address.city)")

	// Authentication Flags
	compareCmd.Flags().StringVar(&basicUser, "basic-username", "",
		"Username for Basic Auth")
	compareCmd.Flags().StringVar(&basicPass, "basic-password", "",
		"Password for Basic Auth. Can use syntax: @file:/path or $ENV_NAME")

	compareCmd.Flags().StringVar(&bearerToken, "bearer-token", "",
		"Bearer token for Auth. Can use syntax: @file:/path or $ENV_NAME")

	// HTTP Headers
	compareCmd.Flags().StringArrayVar(&customHeaders, "header", nil,
		"Add custom HTTP headers (e.g., 'X-API-Key=mykey')")

	_ = compareCmd.MarkFlagFilename("basic-password") // optional UX hint
}

// resolveValue resolves value from env var or file
func resolveValue(val string) string {
	if val == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(val, "$"):
		return os.Getenv(strings.TrimPrefix(val, "$"))
	case strings.HasPrefix(val, "@"):
		path := strings.TrimPrefix(val, "@")
		content, _ := os.ReadFile(path)
		return strings.TrimSpace(string(content))
	default:
		return val
	}
}

// countLines counts number of lines in diff output
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n")) - 1
}
