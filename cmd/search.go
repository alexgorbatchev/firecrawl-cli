package cmd

import (
	"encoding/json"
	"fmt"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
	"github.com/spf13/cobra"
)

var (
	// Local search flag variables
	searchIncludeDomains        []string
	searchExcludeDomains        []string
	searchLimit                 int
	searchTBS                   string
	searchLocation              string
	searchIgnoreInvalidURLs     bool
	searchScrapeFormats         []string
	searchScrapeOnlyMainContent bool
)

var searchCmd = &cobra.Command{
	Use:   "search [QUERY]",
	Short: "Search the web and get scraped results",
	Long:  `Search the web and return scraped content for matching results using the Firecrawl API.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		client, err := getClient()
		if err != nil {
			return err
		}

		opts := &firecrawl.SearchOptions{}

		if cmd.Flags().Changed("include-domains") {
			opts.IncludeDomains = searchIncludeDomains
		}
		if cmd.Flags().Changed("exclude-domains") {
			opts.ExcludeDomains = searchExcludeDomains
		}
		if cmd.Flags().Changed("limit") {
			opts.Limit = firecrawl.Int(searchLimit)
		}
		if cmd.Flags().Changed("tbs") {
			opts.TBS = firecrawl.String(searchTBS)
		}
		if cmd.Flags().Changed("location") {
			opts.Location = firecrawl.String(searchLocation)
		}
		if cmd.Flags().Changed("ignore-invalid-urls") {
			opts.IgnoreInvalidURLs = firecrawl.Bool(searchIgnoreInvalidURLs)
		}

		// Configure embedded scrape options if requested
		if cmd.Flags().Changed("scrape-formats") || cmd.Flags().Changed("scrape-only-main-content") {
			opts.ScrapeOptions = &firecrawl.ScrapeOptions{}
			if cmd.Flags().Changed("scrape-formats") {
				opts.ScrapeOptions.Formats = searchScrapeFormats
			} else {
				opts.ScrapeOptions.Formats = []string{"markdown"}
			}
			if cmd.Flags().Changed("scrape-only-main-content") {
				opts.ScrapeOptions.OnlyMainContent = firecrawl.Bool(searchScrapeOnlyMainContent)
			}
		}

		// Run search operation
		searchData, err := client.Search(cmd.Context(), query, opts)
		if err != nil {
			return fmt.Errorf("searching failed: %w", err)
		}

		// Output result
		if jsonOutput {
			bz, err := json.MarshalIndent(searchData, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling search data: %w", err)
			}
			cmd.Println(string(bz))
			return nil
		}

		// Human-friendly output
		if len(searchData.Web) == 0 && len(searchData.News) == 0 && len(searchData.Images) == 0 {
			cmd.Println("No search results found.")
			return nil
		}

		printResultCategory := func(name string, items []map[string]interface{}) {
			if len(items) == 0 {
				return
			}
			cmd.Printf("=== %s Results ===\n\n", name)
			for i, item := range items {
				title, _ := item["title"].(string)
				url, _ := item["url"].(string)
				description, _ := item["description"].(string)
				if description == "" {
					description, _ = item["snippet"].(string)
				}

				cmd.Printf("[%d] %s\n", i+1, title)
				cmd.Printf("    URL: %s\n", url)
				if description != "" {
					cmd.Printf("    Snippet: %s\n", description)
				}
				if markdown, ok := item["markdown"].(string); ok && markdown != "" {
					cmd.Printf("    Markdown Length: %d characters\n", len(markdown))
				}
				cmd.Println()
			}
		}

		printResultCategory("Web", searchData.Web)
		printResultCategory("News", searchData.News)
		printResultCategory("Image", searchData.Images)

		return nil
	},
}

func init() {
	// Register flags for search command - NO shorthand single-character flags (only double-dash)
	searchCmd.Flags().StringSliceVar(&searchIncludeDomains, "include-domains", nil, "Domains to restrict the search to")
	searchCmd.Flags().StringSliceVar(&searchExcludeDomains, "exclude-domains", nil, "Domains to exclude from the search")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 5, "Maximum number of search results to return")
	searchCmd.Flags().StringVar(&searchTBS, "tbs", "", "Time-based search restriction (e.g. qdr:d for past 24h, qdr:w for past week)")
	searchCmd.Flags().StringVar(&searchLocation, "location", "", "Location config for search (e.g. 'United States' or geolocation JSON)")
	searchCmd.Flags().BoolVar(&searchIgnoreInvalidURLs, "ignore-invalid-urls", false, "Ignore invalid URLs found in search results")

	// Embedded scrape options
	searchCmd.Flags().StringSliceVar(&searchScrapeFormats, "scrape-formats", []string{"markdown"}, "Formats for scraping matching pages (e.g. markdown, html)")
	searchCmd.Flags().BoolVar(&searchScrapeOnlyMainContent, "scrape-only-main-content", true, "Only return main content of matching pages during scrape")

	RootCmd.AddCommand(searchCmd)
}
