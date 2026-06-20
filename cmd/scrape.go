package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// Local scrape flag variables
	scrapeFormat              []string
	scrapeHtml                bool
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
	scrapeSchema              string
	scrapeSchemaFile          string
	scrapeActions             string
	scrapeActionsFile         string
	scrapeScreenshot          bool
	scrapeFullPageScreenshot  bool
	scrapeOutput              string
	scrapePretty              bool
	scrapeTiming              bool
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

		// Map formats
		formats := []string{}
		if cmd.Flags().Changed("format") {
			formats = scrapeFormat
		} else if scrapeHtml {
			formats = []string{"html"}
		} else {
			formats = []string{"markdown"}
		}

		if scrapeScreenshot && !contains(formats, "screenshot") {
			formats = append(formats, "screenshot")
		}
		opts.Formats = formats

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

		// Handle actions
		if scrapeActions != "" || scrapeActionsFile != "" {
			var actionsList []map[string]interface{}
			var actionsBytes []byte
			var fileErr error

			if scrapeActionsFile != "" {
				actionsBytes, fileErr = os.ReadFile(scrapeActionsFile)
				if fileErr != nil {
					return fmt.Errorf("reading actions file: %w", fileErr)
				}
			} else {
				actionsBytes = []byte(scrapeActions)
			}

			if err := json.Unmarshal(actionsBytes, &actionsList); err != nil {
				return fmt.Errorf("parsing actions JSON: %w", err)
			}
			opts.Actions = actionsList
		}

		// Handle schema for structured JSON Options
		if scrapeJsonPrompt != "" || scrapeSchema != "" || scrapeSchemaFile != "" {
			opts.JsonOptions = &firecrawl.JsonOptions{}
			if scrapeJsonPrompt != "" {
				opts.JsonOptions.Prompt = scrapeJsonPrompt
			}
			if scrapeSchema != "" || scrapeSchemaFile != "" {
				var schemaMap map[string]interface{}
				var schemaBytes []byte
				var schemaErr error

				if scrapeSchemaFile != "" {
					schemaBytes, schemaErr = os.ReadFile(scrapeSchemaFile)
					if schemaErr != nil {
						return fmt.Errorf("reading schema file: %w", schemaErr)
					}
				} else {
					schemaBytes = []byte(scrapeSchema)
				}

				if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
					return fmt.Errorf("parsing schema JSON: %w", err)
				}
				opts.JsonOptions.Schema = schemaMap
			}
		}

		// Track timing if requested
		startTime := time.Now()

		// Run the scrape operation
		doc, err := client.Scrape(cmd.Context(), url, opts)
		if err != nil {
			return fmt.Errorf("scraping failed: %w", err)
		}

		duration := time.Since(startTime)

		// Format and output the result
		var outputStr string

		if jsonOutput || len(formats) > 1 {
			var bz []byte
			var mErr error
			if scrapePretty || jsonOutput {
				bz, mErr = json.MarshalIndent(doc, "", "  ")
			} else {
				bz, mErr = json.Marshal(doc)
			}
			if mErr != nil {
				return fmt.Errorf("marshaling document: %w", mErr)
			}
			outputStr = string(bz)
		} else {
			// Single format output
			if doc.Markdown != "" {
				outputStr = doc.Markdown
			} else if doc.JSON != nil {
				bz, err := json.MarshalIndent(doc.JSON, "", "  ")
				if err == nil {
					outputStr = string(bz)
				} else {
					outputStr = fmt.Sprintf("%+v", doc.JSON)
				}
			} else if doc.HTML != "" {
				outputStr = doc.HTML
			} else if doc.RawHTML != "" {
				outputStr = doc.RawHTML
			} else {
				// Fallback: dump metadata
				bz, err := json.MarshalIndent(doc.Metadata, "", "  ")
				if err == nil {
					outputStr = string(bz)
				} else {
					outputStr = "Scraped successfully, but no markdown or HTML content was returned."
				}
			}
		}

		// If timing requested, append to output or stderr
		if scrapeTiming {
			timingInfo := fmt.Sprintf("\n--- Request Timing ---\nDuration: %v\n", duration)
			if scrapeOutput != "" {
				cmd.PrintErr(timingInfo)
			} else {
				outputStr += timingInfo
			}
		}

		// Save to file or write to stdout
		if scrapeOutput != "" {
			err := os.WriteFile(scrapeOutput, []byte(outputStr), 0644)
			if err != nil {
				return fmt.Errorf("writing output to file: %w", err)
			}
		} else {
			cmd.Println(outputStr)
		}

		return nil
	},
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func registerScrapeFlags(f *pflag.FlagSet) {
	// Register flags for scrape - NO shorthand single-character flags (only double-dash)
	f.StringSliceVar(&scrapeFormat, "format", []string{"markdown"}, "Output formats (comma-separated): markdown, html, rawHtml, links, screenshot, json, images, summary, changeTracking, attributes, branding")
	f.BoolVar(&scrapeHtml, "html", false, "Shortcut for --format html")
	f.BoolVar(&scrapeOnlyMainContent, "only-main-content", true, "Extract only main content")
	f.StringSliceVar(&scrapeIncludeTags, "include-tags", nil, "HTML tags to include (comma-separated)")
	f.StringSliceVar(&scrapeExcludeTags, "exclude-tags", nil, "HTML tags to exclude (comma-separated)")
	f.IntVar(&scrapeWaitFor, "wait-for", 0, "Wait time in milliseconds for JS rendering")
	f.BoolVar(&scrapeMobile, "mobile", false, "Enable mobile user-agent scraping")
	f.BoolVar(&scrapeSkipTLSVerification, "skip-tls-verification", false, "Skip TLS certificate verification")
	f.BoolVar(&scrapeRemoveBase64Images, "remove-base64-images", false, "Remove base64-encoded images from the output")
	f.BoolVar(&scrapeBlockAds, "block-ads", false, "Block advertisement elements")
	f.StringVar(&scrapeProxy, "proxy", "", "Proxy mode for scraping (e.g., auto or basic)")
	f.Int64Var(&scrapeMaxAge, "max-age", 0, "Maximum cache age in seconds")
	f.BoolVar(&scrapeStoreInCache, "store-in-cache", false, "Store scraped content in cache")
	f.BoolVar(&scrapeLockdown, "lockdown", false, "Enable strict lockdown mode")
	f.BoolVar(&scrapeRedactPII, "redact-pii", false, "Redact personally identifiable information")
	f.StringVar(&scrapeLocationCountry, "location-country", "", "ISO country code for geotargeting (e.g. US, DE)")
	f.StringSliceVar(&scrapeLocationLanguages, "location-languages", nil, "Languages to request for geotargeting (e.g. en, fr)")
	f.StringVar(&scrapeJsonPrompt, "json-prompt", "", "Prompt for structural JSON extraction")
	f.StringVar(&scrapeSchema, "schema", "", "JSON schema for structured extraction (inline JSON string)")
	f.StringVar(&scrapeSchemaFile, "schema-file", "", "Path to JSON schema file")
	f.StringVar(&scrapeActions, "actions", "", "JSON actions array to run during scrape (inline JSON)")
	f.StringVar(&scrapeActionsFile, "actions-file", "", "Path to JSON actions file")
	f.BoolVar(&scrapeScreenshot, "screenshot", false, "Take a screenshot")
	f.BoolVar(&scrapeFullPageScreenshot, "full-page-screenshot", false, "Take a full page screenshot")
	f.StringVar(&scrapeOutput, "output", "", "Save output to file")
	f.BoolVar(&scrapePretty, "pretty", false, "Pretty print JSON output")
	f.BoolVar(&scrapeTiming, "timing", false, "Show request timing and other useful information")
}

func init() {
	registerScrapeFlags(scrapeCmd.Flags())
	RootCmd.AddCommand(scrapeCmd)
}
