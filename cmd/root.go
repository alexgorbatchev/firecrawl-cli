package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Global flag variables
	apiKey     string
	apiURL     string
	timeout    time.Duration
	jsonOutput bool

	// Version represents the CLI version, injected at build time by GoReleaser.
	Version = "dev"

	// clientFactory defines how we create the Firecrawl client.
	// We make it a variable so we can mock it in tests.
	createClient = func(apiKey, apiURL string, timeout time.Duration) (FirecrawlClient, error) {
		return NewRealFirecrawlClient(apiKey, apiURL, timeout)
	}
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:           "firecrawl [URL]",
	Version:       Version,
	Short:         "A utility for web scraping, mapping, searching, and AI-agent browser tasks.",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `firecrawl is a command-line interface for crawling, scraping, and extracting structured web data.
It maps URLs, scrapes single pages, searches the web, and runs AI-powered agent extraction tasks.

You can customize the Firecrawl API URL with --api-url.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If 1 argument starting with http:// or https://, route to scrape command
		if len(args) == 1 && (strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://")) {
			return scrapeCmd.RunE(cmd, args)
		}
		// Otherwise print help
		return cmd.Help()
	},
}

func init() {
	// Root persistent flags (available to all subcommands)
	// We only use double dash arguments, so we do not define short flags.
	RootCmd.PersistentFlags().StringVar(&apiKey, "api-key", os.Getenv("FIRECRAWL_API_KEY"), "Firecrawl API key (defaults to FIRECRAWL_API_KEY env var)")
	RootCmd.PersistentFlags().StringVar(&apiURL, "api-url", os.Getenv("FIRECRAWL_API_URL"), "Firecrawl API base URL (defaults to FIRECRAWL_API_URL env var)")
	RootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 5*time.Minute, "Timeout for the API client operations (e.g., 30s, 5m)")
	RootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results as raw JSON instead of human-friendly formatting")

	// Define a custom help flag to avoid Cobra's default short -h flag, ensuring only double-dash flags exist.
	RootCmd.PersistentFlags().Bool("help", false, "help for firecrawl")

	// Register local scrape flags on the root command so they are accepted when running direct URL scrapes
	registerScrapeFlags(RootCmd.Flags())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// getClient is a helper to instantiate the FirecrawlClient using the parsed flags.
func getClient() (FirecrawlClient, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_URL or --api-url must be set.")
	}
	return createClient(apiKey, apiURL, timeout)
}
