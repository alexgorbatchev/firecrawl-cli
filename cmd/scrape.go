package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
	"github.com/spf13/cobra"
)

var (
	// Local scrape flag variables
	scrapeFormats             []string
	scrapeOnlyMainContent     bool
	scrapeIncludeTags         []string
	scrapeExcludeTags         []string
	scrapeWaitFor             int
	scrapeMobile              bool
	scrapeSkipTLSVerification bool
	scrapeRemoveBase64Images  bool
	scrapeBlockAds            bool
	scrapeProxy               string
	scrapeMaxAge              int64
	scrapeStoreInCache        bool
	scrapeLockdown            bool
	scrapeRedactPII           bool
	scrapeLocationCountry     string
	scrapeLocationLanguages   []string
	scrapeJsonPrompt          string
	scrapeJsonSchema          string
)

var scrapeCmd = &cobra.Command{
	Use:   "scrape [URL]",
	Short: "Scrape a single page and get its content",
	Long:  `Scrape a single URL using the Firecrawl API and return its structured or unstructured content (e.g., markdown, HTML, json).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		client, err := getClient()
		if err != nil {
			return err
		}

		opts := &firecrawl.ScrapeOptions{}

		// Map flag changes to ScrapeOptions pointers dynamically
		if cmd.Flags().Changed("formats") {
			opts.Formats = scrapeFormats
		} else {
			// Default to markdown if not specified
			opts.Formats = []string{"markdown"}
		}

		if cmd.Flags().Changed("only-main-content") {
			opts.OnlyMainContent = firecrawl.Bool(scrapeOnlyMainContent)
		}
		if cmd.Flags().Changed("include-tags") {
			opts.IncludeTags = scrapeIncludeTags
		}
		if cmd.Flags().Changed("exclude-tags") {
			opts.ExcludeTags = scrapeExcludeTags
		}
		if cmd.Flags().Changed("wait-for") {
			opts.WaitFor = firecrawl.Int(scrapeWaitFor)
		}
		if cmd.Flags().Changed("mobile") {
			opts.Mobile = firecrawl.Bool(scrapeMobile)
		}
		if cmd.Flags().Changed("skip-tls-verification") {
			opts.SkipTLSVerification = firecrawl.Bool(scrapeSkipTLSVerification)
		}
		if cmd.Flags().Changed("remove-base64-images") {
			opts.RemoveBase64Images = firecrawl.Bool(scrapeRemoveBase64Images)
		}
		if cmd.Flags().Changed("block-ads") {
			opts.BlockAds = firecrawl.Bool(scrapeBlockAds)
		}
		if cmd.Flags().Changed("proxy") {
			opts.Proxy = firecrawl.String(scrapeProxy)
		}
		if cmd.Flags().Changed("max-age") {
			opts.MaxAge = firecrawl.Int64(scrapeMaxAge)
		}
		if cmd.Flags().Changed("store-in-cache") {
			opts.StoreInCache = firecrawl.Bool(scrapeStoreInCache)
		}
		if cmd.Flags().Changed("lockdown") {
			opts.Lockdown = firecrawl.Bool(scrapeLockdown)
		}
		if cmd.Flags().Changed("redact-pii") {
			opts.RedactPII = firecrawl.Bool(scrapeRedactPII)
		}

		// Geolocation targeting configuration
		if scrapeLocationCountry != "" || len(scrapeLocationLanguages) > 0 {
			opts.Location = &firecrawl.LocationConfig{}
			if scrapeLocationCountry != "" {
				opts.Location.Country = scrapeLocationCountry
			}
			if len(scrapeLocationLanguages) > 0 {
				opts.Location.Languages = scrapeLocationLanguages
			}
		}

		// Handle structured extraction JSON Options
		if scrapeJsonPrompt != "" || scrapeJsonSchema != "" {
			opts.JsonOptions = &firecrawl.JsonOptions{}
			if scrapeJsonPrompt != "" {
				opts.JsonOptions.Prompt = scrapeJsonPrompt
			}
			if scrapeJsonSchema != "" {
				var schemaMap map[string]interface{}
				// Try parsing as raw JSON string first
				if err := json.Unmarshal([]byte(scrapeJsonSchema), &schemaMap); err != nil {
					// If parsing fails, try reading as a file path
					fileBytes, fileErr := os.ReadFile(scrapeJsonSchema)
					if fileErr != nil {
						return fmt.Errorf("json-schema is neither valid JSON nor a readable file path: %w (json parse err: %v)", fileErr, err)
					}
					if err := json.Unmarshal(fileBytes, &schemaMap); err != nil {
						return fmt.Errorf("parsing JSON schema from file: %w", err)
					}
				}
				opts.JsonOptions.Schema = schemaMap
			}
		}

		// Run the scrape operation
		doc, err := client.Scrape(cmd.Context(), url, opts)
		if err != nil {
			return fmt.Errorf("scraping failed: %w", err)
		}

		// Format and output the result
		if jsonOutput {
			bz, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling document: %w", err)
			}
			cmd.Println(string(bz))
			return nil
		}

		// Human-friendly output (prefer markdown, then JSON/Html/Text)
		if doc.Markdown != "" {
			cmd.Println(doc.Markdown)
		} else if doc.JSON != nil {
			bz, err := json.MarshalIndent(doc.JSON, "", "  ")
			if err == nil {
				cmd.Println(string(bz))
			} else {
				cmd.Printf("%+v\n", doc.JSON)
			}
		} else if doc.HTML != "" {
			cmd.Println(doc.HTML)
		} else if doc.RawHTML != "" {
			cmd.Println(doc.RawHTML)
		} else {
			// Fallback: dump metadata
			bz, err := json.MarshalIndent(doc.Metadata, "", "  ")
			if err == nil {
				cmd.Println(string(bz))
			} else {
				cmd.Println("Scraped successfully, but no markdown or HTML content was returned.")
			}
		}

		return nil
	},
}

func init() {
	// Register flags for scrape command - NO shorthand single-character flags (only double-dash)
	scrapeCmd.Flags().StringSliceVar(&scrapeFormats, "formats", []string{"markdown"}, "Formats to return (e.g. markdown, html, rawHtml, screenshot, links, video, product, json)")
	scrapeCmd.Flags().BoolVar(&scrapeOnlyMainContent, "only-main-content", true, "Only return main content of the page")
	scrapeCmd.Flags().StringSliceVar(&scrapeIncludeTags, "include-tags", nil, "Comma-separated HTML tags to include")
	scrapeCmd.Flags().StringSliceVar(&scrapeExcludeTags, "exclude-tags", nil, "Comma-separated HTML tags to exclude")
	scrapeCmd.Flags().IntVar(&scrapeWaitFor, "wait-for", 0, "Wait time in milliseconds before scraping")
	scrapeCmd.Flags().BoolVar(&scrapeMobile, "mobile", false, "Enable mobile user-agent scraping")
	scrapeCmd.Flags().BoolVar(&scrapeSkipTLSVerification, "skip-tls-verification", false, "Skip TLS certificate verification")
	scrapeCmd.Flags().BoolVar(&scrapeRemoveBase64Images, "remove-base64-images", false, "Remove base64-encoded images from the output")
	scrapeCmd.Flags().BoolVar(&scrapeBlockAds, "block-ads", false, "Block advertisement elements")
	scrapeCmd.Flags().StringVar(&scrapeProxy, "proxy", "", "Proxy server URL (e.g., http://proxy.example.com:8080)")
	scrapeCmd.Flags().Int64Var(&scrapeMaxAge, "max-age", 0, "Maximum cache age in seconds")
	scrapeCmd.Flags().BoolVar(&scrapeStoreInCache, "store-in-cache", false, "Store scraped content in cache")
	scrapeCmd.Flags().BoolVar(&scrapeLockdown, "lockdown", false, "Enable strict lockdown mode")
	scrapeCmd.Flags().BoolVar(&scrapeRedactPII, "redact-pii", false, "Redact personally identifiable information")
	scrapeCmd.Flags().StringVar(&scrapeLocationCountry, "location-country", "", "ISO country code for geotargeting (e.g. US, DE)")
	scrapeCmd.Flags().StringSliceVar(&scrapeLocationLanguages, "location-languages", nil, "Languages to request for geotargeting (e.g. en, fr)")
	scrapeCmd.Flags().StringVar(&scrapeJsonPrompt, "json-prompt", "", "Prompt for structural JSON extraction")
	scrapeCmd.Flags().StringVar(&scrapeJsonSchema, "json-schema", "", "Raw JSON schema string or path to a JSON schema file")

	RootCmd.AddCommand(scrapeCmd)
}
