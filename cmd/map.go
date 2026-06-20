package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
	"github.com/spf13/cobra"
)

var (
	// Local map flag variables
	mapSearch                string
	mapSitemap               string
	mapIncludeSubdomains     bool
	mapIgnoreQueryParameters bool
	mapLimit                 int
	mapLocationCountry       string
	mapLocationLanguages     []string
	mapDetailed              bool
	mapWait                  bool
	mapTimeout               int
	mapOutput                string
	mapPretty                bool
)

var mapCmd = &cobra.Command{
	Use:   "map [URL]",
	Short: "Discover all URLs on a website quickly",
	Long:  `Discover and map URLs on a website starting from a given root URL using the Firecrawl API.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		client, err := getClient()
		if err != nil {
			return err
		}

		opts := &firecrawl.MapOptions{}

		if cmd.Flags().Changed("search") {
			opts.Search = firecrawl.String(mapSearch)
		}
		if cmd.Flags().Changed("sitemap") {
			opts.Sitemap = firecrawl.String(mapSitemap)
		}
		if cmd.Flags().Changed("include-subdomains") {
			opts.IncludeSubdomains = firecrawl.Bool(mapIncludeSubdomains)
		}
		if cmd.Flags().Changed("ignore-query-parameters") {
			opts.IgnoreQueryParameters = firecrawl.Bool(mapIgnoreQueryParameters)
		}
		if cmd.Flags().Changed("limit") {
			opts.Limit = firecrawl.Int(mapLimit)
		}
		if cmd.Flags().Changed("timeout") {
			opts.Timeout = firecrawl.Int(mapTimeout)
		}

		// Geolocation targeting configuration
		if mapLocationCountry != "" || len(mapLocationLanguages) > 0 {
			opts.Location = &firecrawl.LocationConfig{}
			if mapLocationCountry != "" {
				opts.Location.Country = mapLocationCountry
			}
			if len(mapLocationLanguages) > 0 {
				opts.Location.Languages = mapLocationLanguages
			}
		}

		// Run map operation
		mapData, err := client.Map(cmd.Context(), url, opts)
		if err != nil {
			return fmt.Errorf("mapping failed: %w", err)
		}

		var outputStr string

		// Output result format
		if jsonOutput {
			var bz []byte
			var mErr error
			if mapPretty {
				bz, mErr = json.MarshalIndent(mapData, "", "  ")
			} else {
				bz, mErr = json.Marshal(mapData)
			}
			if mErr != nil {
				return fmt.Errorf("marshaling map data: %w", mErr)
			}
			outputStr = string(bz)
		} else {
			if len(mapData.Links) == 0 {
				outputStr = "No URLs discovered."
			} else {
				for i, link := range mapData.Links {
					line := ""
					if mapDetailed {
						titleStr := ""
						descStr := ""
						if link.Title != "" {
							titleStr = fmt.Sprintf(" (Title: %s)", link.Title)
						}
						if link.Description != "" {
							descStr = fmt.Sprintf(" - %s", link.Description)
						}
						line = fmt.Sprintf("%s%s%s", link.URL, titleStr, descStr)
					} else {
						line = link.URL
					}
					if i > 0 {
						outputStr += "\n"
					}
					outputStr += line
				}
			}
		}

		// Write output to file if requested, otherwise print to stdout
		if mapOutput != "" {
			err := os.WriteFile(mapOutput, []byte(outputStr), 0644)
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
	// Register flags for map command - NO shorthand single-character flags (only double-dash)
	mapCmd.Flags().StringVar(&mapSearch, "search", "", "Filter URLs by search query")
	mapCmd.Flags().StringVar(&mapSitemap, "sitemap", "", "Sitemap handling: include, skip, only, or custom sitemap URL")
	mapCmd.Flags().BoolVar(&mapIncludeSubdomains, "include-subdomains", false, "Include subdomains in mapped URLs")
	mapCmd.Flags().BoolVar(&mapIgnoreQueryParameters, "ignore-query-parameters", false, "Treat URLs with different params as same")
	mapCmd.Flags().IntVar(&mapLimit, "limit", 100, "Maximum URLs to discover")
	mapCmd.Flags().StringVar(&mapLocationCountry, "location-country", "", "ISO country code for geotargeting (e.g. US, DE)")
	mapCmd.Flags().StringSliceVar(&mapLocationLanguages, "location-languages", nil, "Languages to request for geotargeting (e.g. en, fr)")
	mapCmd.Flags().BoolVar(&mapDetailed, "detailed", false, "Show details like title and description for discovered links")
	mapCmd.Flags().BoolVar(&mapWait, "wait", false, "Wait for map to complete")
	mapCmd.Flags().IntVar(&mapTimeout, "timeout", 0, "Timeout in seconds")
	mapCmd.Flags().StringVar(&mapOutput, "output", "", "Save output to file")
	mapCmd.Flags().BoolVar(&mapPretty, "pretty", false, "Pretty print JSON output")

	RootCmd.AddCommand(mapCmd)
}
