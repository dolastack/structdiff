// cmd/compare.go
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dolastack/structdiff/internal/parser"
	"github.com/spf13/cobra"
)

var (
	checkFlag         bool
	quietFlag         bool
	filterPaths       []string
	basicUser         string
	basicPass         string
	bearerToken       string
	oauthClientID     string
	oauthClientSecret string
	oauthTokenURL     string

	ssoEnabled  bool
	ssoClientID string
	ssoTokenURL string
	ssoScopes   []string

	awsSigv4      string
	awsAssumeRole string
	customHeaders []string

	tokenCacheEnabled bool
	tokenCachePath    string
)

var compareCmd = &cobra.Command{
	Use:     "compare [file1] [file2]",
	Short:   "Compare two structured files (JSON, YAML, TOML, XML, INI, CSV)",
	Example: "structdiff compare file1.yaml file2.yaml --filter=user.name",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		file1 := args[0]
		file2 := args[1]

		// Set auth
		parser.SetAuth(basicUser, basicPass, bearerToken)
		parser.SetCustomHeaders(customHeaders)

		// OAuth2
		if oauthClientID != "" && oauthClientSecret != "" && oauthTokenURL != "" {
			err := parser.SetOAuthConfig(oauthClientID, oauthClientSecret, oauthTokenURL)
			if err != nil {
				return err
			}
		}

		// SSO / Device Flow
		if ssoEnabled {
			if ssoClientID == "" || ssoTokenURL == "" {
				return fmt.Errorf("--sso requires --sso-client-id and --sso-token-url")
			}
			err := parser.SetDeviceFlowConfig(ssoClientID, ssoTokenURL, ssoScopes)
			if err != nil {
				return err
			}
		}

		// AWS Sigv4
		if awsSigv4 != "" {
			parts := strings.Split(awsSigv4, ";")
			service := ""
			region := ""

			for _, p := range parts {
				kv := strings.SplitN(p, ":", 2)
				if len(kv) == 2 {
					switch kv[0] {
					case "service":
						service = strings.TrimSpace(kv[1])
					case "region":
						region = strings.TrimSpace(kv[1])
					}
				}
			}

			if service == "" || region == "" {
				return fmt.Errorf("AWS Sigv4 requires service and region")
			}

			if err := parser.SetAwsSigv4(service, region, awsAssumeRole); err != nil {
				return err
			}
		}

		// Token caching
		if tokenCacheEnabled {
			parser.SetTokenCache(tokenCachePath)
			parser.EnableTokenCache(true)
		}

		diffOutput, err := diff.CompareFiles(file1, file2, quietFlag, filterPaths)
		if err != nil {
			return fmt.Errorf("error comparing files: %v", err)
		}

		if diffOutput != "" && !quietFlag {
			fmt.Print(diffOutput)
		}

		if checkFlag && strings.Contains(diffOutput, "Found") {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)

	// Auth flags
	compareCmd.Flags().BoolVarP(&checkFlag, "check", "c", false,
		"Exit with non-zero code if differences found")
	compareCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false,
		"Only output summary (use with --check)")
	compareCmd.Flags().StringArrayVarP(&filterPaths, "filter", "f", nil,
		"Only show diffs under this key path (e.g., user.address.city)")

	// Basic Auth
	compareCmd.Flags().StringVar(&basicUser, "basic-username", "", "Username for Basic Auth")
	compareCmd.Flags().StringVar(&basicPass, "basic-password", "", "Password for Basic Auth")

	// Bearer Token
	compareCmd.Flags().StringVar(&bearerToken, "bearer-token", "",
		"Bearer token for Auth. Can also be read from a file or env var using syntax: @file:/path or $ENV_NAME")

	// OAuth2
	compareCmd.Flags().StringVar(&oauthClientID, "oauth-client-id", "", "OAuth2 client ID for token fetch")
	compareCmd.Flags().StringVar(&oauthClientSecret, "oauth-client-secret", "", "OAuth2 client secret for token fetch")
	compareCmd.Flags().StringVar(&oauthTokenURL, "oauth-token-url", "", "OAuth2 token URL for client credentials flow")

	// SSO
	compareCmd.Flags().BoolVar(&ssoEnabled, "sso", false, "Use interactive SSO login (device code flow)")
	compareCmd.Flags().StringVar(&ssoClientID, "sso-client-id", "", "OAuth2 client ID for SSO")
	compareCmd.Flags().StringVar(&ssoTokenURL, "sso-token-url", "", "OAuth2 token URL for SSO")
	compareCmd.Flags().StringArrayVar(&ssoScopes, "sso-scope", []string{"openid", "profile"},
		"OAuth2 scopes for SSO authentication")

	// AWS Sigv4
	compareCmd.Flags().StringVar(&awsSigv4, "aws-sigv4", "",
		"AWS Sigv4 signing config. Format: 'service:name;region:region'")
	compareCmd.Flags().StringVar(&awsAssumeRole, "aws-assume-role", "",
		"AWS IAM role ARN to assume. Requires AWS Sigv4 mode (--aws-sigv4)")

	// Custom Headers
	compareCmd.Flags().StringArrayVar(&customHeaders, "header", nil,
		"Add custom HTTP headers (e.g., 'X-API-Key=mykey')")

	// Token Caching
	compareCmd.Flags().BoolVar(&tokenCacheEnabled, "token-cache", false,
		"Enable session token caching")
	compareCmd.Flags().StringVar(&tokenCachePath, "token-path", "",
		"Custom path to store tokens (default: ~/.structdiff/token.json")
}
