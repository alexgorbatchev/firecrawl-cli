package cmd

import (
	"encoding/json"
	"fmt"

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
)

var mapCmd = &cobra.Command{
	Use:   "map [URL]",
	Short: "Discover URLs on a website",
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

		// Output result
		if jsonOutput {
			bz, err := json.MarshalIndent(mapData, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling map data: %w", err)
			}
			cmd.Println(string(bz))
			return nil
		}

		if len(mapData.Links) == 0 {
			cmd.Println("No URLs discovered.")
			return nil
		}

		for _, link := range mapData.Links {
			if mapDetailed {
				titleStr := ""
				descStr := ""
				if link.Title != "" {
					titleStr = fmt.Sprintf(" (Title: %s)", link.Title)
				}
				if link.Description != "" {
					descStr = fmt.Sprintf(" - %s", link.Description)
				}
				cmd.Printf("%s%s%s\n", link.URL, titleStr, descStr)
			} else {
				cmd.Println(link.URL)
			}
		}

		return nil
	},
}

func init() {
	// Register flags for map command - NO shorthand single-character flags (only double-dash)
	mapCmd.Flags().StringVar(&mapSearch, "search", "", "Search query to filter discovered URLs")
	mapCmd.Flags().StringVar(&mapSitemap, "sitemap", "", "Custom sitemap XML URL to use for discovery")
	mapCmd.Flags().BoolVar(&mapIncludeSubdomains, "include-subdomains", false, "Include subdomains of the main URL")
	mapCmd.Flags().BoolVar(&mapIgnoreQueryParameters, "ignore-query-parameters", false, "Ignore query parameters in discovered URLs")
	mapCmd.Flags().IntVar(&mapLimit, "limit", 100, "Maximum number of links to return")
	mapCmd.Flags().StringVar(&mapLocationCountry, "location-country", "", "ISO country code for geotargeting (e.g. US, DE)")
	mapCmd.Flags().StringSliceVar(&mapLocationLanguages, "location-languages", nil, "Languages to request for geotargeting (e.g. en, fr)")
	mapCmd.Flags().BoolVar(&mapDetailed, "detailed", false, "Show details like title and description for discovered links")

	RootCmd.AddCommand(mapCmd)
}
