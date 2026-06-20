package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
	"github.com/spf13/cobra"
)

var (
	// Local search flag variables
	searchLimit                 int
	searchSources               []string
	searchCategories            []string
	searchTBS                   string
	searchLocation              string
	searchCountry               string
	searchTimeout               int
	searchIgnoreInvalidURLs     bool
	searchScrape                bool
	searchScrapeFormats         []string
	searchScrapeOnlyMainContent bool
	searchOutput                string
	searchPretty                bool
)

var searchCmd = &cobra.Command{
	Use:   "search [QUERY]",
	Short: "Search the web and optionally scrape the results",
	Long:  `Search the web and return scraped content for matching results using the Firecrawl API.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		client, err := getClient()
		if err != nil {
			return err
		}

		opts := &firecrawl.SearchOptions{}

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

		// Handle sources mapping
		if cmd.Flags().Changed("sources") {
			opts.Sources = make([]interface{}, len(searchSources))
			for i, v := range searchSources {
				opts.Sources[i] = v
			}
		}

		// Handle categories mapping
		if cmd.Flags().Changed("categories") {
			opts.Categories = make([]interface{}, len(searchCategories))
			for i, v := range searchCategories {
				opts.Categories[i] = v
			}
		}

		// Configure scrape options if --scrape is specified
		if searchScrape {
			opts.ScrapeOptions = &firecrawl.ScrapeOptions{}
			if cmd.Flags().Changed("scrape-formats") {
				opts.ScrapeOptions.Formats = searchScrapeFormats
			} else {
				opts.ScrapeOptions.Formats = []string{"markdown"}
			}
			opts.ScrapeOptions.OnlyMainContent = firecrawl.Bool(searchScrapeOnlyMainContent)
		}

		// Run search operation
		searchData, err := client.Search(cmd.Context(), query, opts)
		if err != nil {
			return fmt.Errorf("searching failed: %w", err)
		}

		var outputStr string

		// Output result handling
		if jsonOutput {
			var bz []byte
			var mErr error
			if searchPretty {
				bz, mErr = json.MarshalIndent(searchData, "", "  ")
			} else {
				bz, mErr = json.Marshal(searchData)
			}
			if mErr != nil {
				return fmt.Errorf("marshaling search data: %w", mErr)
			}
			outputStr = string(bz)
		} else {
			if len(searchData.Web) == 0 && len(searchData.News) == 0 && len(searchData.Images) == 0 {
				outputStr = "No search results found."
			} else {
				printResultCategory := func(name string, items []map[string]interface{}) string {
					if len(items) == 0 {
						return ""
					}
					catStr := fmt.Sprintf("=== %s Results ===\n\n", name)
					for i, item := range items {
						title, _ := item["title"].(string)
						url, _ := item["url"].(string)
						description, _ := item["description"].(string)
						if description == "" {
							description, _ = item["snippet"].(string)
						}

						catStr += fmt.Sprintf("[%d] %s\n", i+1, title)
						catStr += fmt.Sprintf("    URL: %s\n", url)
						if description != "" {
							catStr += fmt.Sprintf("    Snippet: %s\n", description)
						}
						if markdown, ok := item["markdown"].(string); ok && markdown != "" {
							catStr += fmt.Sprintf("    Markdown Length: %d characters\n", len(markdown))
						}
						catStr += "\n"
					}
					return catStr
				}

				outputStr += printResultCategory("Web", searchData.Web)
				outputStr += printResultCategory("News", searchData.News)
				outputStr += printResultCategory("Image", searchData.Images)
			}
		}

		// Save output to file or write to stdout
		if searchOutput != "" {
			err := os.WriteFile(searchOutput, []byte(outputStr), 0644)
			if err != nil {
				return fmt.Errorf("writing output to file: %w", err)
			}
		} else {
			cmd.Println(outputStr)
		}

		return nil
	},
}

func init() {
	// Register flags for search command - NO shorthand single-character flags (only double-dash)
	searchCmd.Flags().IntVar(&searchLimit, "limit", 5, "Maximum number of search results to return (max: 100)")
	searchCmd.Flags().StringSliceVar(&searchSources, "sources", nil, "Sources to search (comma-separated): web, images, news")
	searchCmd.Flags().StringSliceVar(&searchCategories, "categories", nil, "Filter by category (comma-separated): github, research, pdf")
	searchCmd.Flags().StringVar(&searchTBS, "tbs", "", "Time-based search restriction (e.g. qdr:d for past day, qdr:w for past week)")
	searchCmd.Flags().StringVar(&searchLocation, "location", "", "Location config for search (e.g. 'Berlin,Germany')")
	searchCmd.Flags().StringVar(&searchCountry, "country", "US", "ISO country code for geotargeting")
	searchCmd.Flags().IntVar(&searchTimeout, "timeout", 60000, "Timeout in milliseconds")
	searchCmd.Flags().BoolVar(&searchIgnoreInvalidURLs, "ignore-invalid-urls", false, "Exclude URLs invalid for other Firecrawl endpoints")
	searchCmd.Flags().BoolVar(&searchScrape, "scrape", false, "Scrape search results")
	searchCmd.Flags().StringSliceVar(&searchScrapeFormats, "scrape-formats", []string{"markdown"}, "Formats for scraped content (comma-separated)")
	searchCmd.Flags().BoolVar(&searchScrapeOnlyMainContent, "scrape-only-main-content", true, "Include only main content when scraping")
	searchCmd.Flags().StringVar(&searchOutput, "output", "", "Save output to file")
	searchCmd.Flags().BoolVar(&searchPretty, "pretty", false, "Pretty print JSON output")

	RootCmd.AddCommand(searchCmd)
}
