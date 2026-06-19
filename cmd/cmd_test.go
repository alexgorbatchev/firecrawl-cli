package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
)

// mockFirecrawlClient implements FirecrawlClient for testing.
type mockFirecrawlClient struct {
	scrapeFn func(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error)
	mapFn    func(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error)
	searchFn func(ctx context.Context, query string, opts *firecrawl.SearchOptions) (*firecrawl.SearchData, error)
	agentFn  func(ctx context.Context, opts *firecrawl.AgentOptions) (*firecrawl.AgentStatusResponse, error)
}

func (m *mockFirecrawlClient) Scrape(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error) {
	if m.scrapeFn != nil {
		return m.scrapeFn(ctx, url, opts)
	}
	return nil, errors.New("Scrape not implemented in mock")
}

func (m *mockFirecrawlClient) Map(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error) {
	if m.mapFn != nil {
		return m.mapFn(ctx, url, opts)
	}
	return nil, errors.New("Map not implemented in mock")
}

func (m *mockFirecrawlClient) Search(ctx context.Context, query string, opts *firecrawl.SearchOptions) (*firecrawl.SearchData, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, opts)
	}
	return nil, errors.New("Search not implemented in mock")
}

func (m *mockFirecrawlClient) Agent(ctx context.Context, opts *firecrawl.AgentOptions) (*firecrawl.AgentStatusResponse, error) {
	if m.agentFn != nil {
		return m.agentFn(ctx, opts)
	}
	return nil, errors.New("Agent not implemented in mock")
}

// resetFlags resets all persistent and local CLI flag variables to their default values.
func resetFlags() {
	// Global
	apiKey = ""
	apiURL = "https://dummy-url.com"
	timeout = 5 * time.Minute
	jsonOutput = false

	// Scrape
	scrapeFormats = []string{"markdown"}
	scrapeOnlyMainContent = true
	scrapeIncludeTags = nil
	scrapeExcludeTags = nil
	scrapeWaitFor = 0
	scrapeMobile = false
	scrapeSkipTLSVerification = false
	scrapeRemoveBase64Images = false
	scrapeBlockAds = false
	scrapeProxy = ""
	scrapeMaxAge = 0
	scrapeStoreInCache = false
	scrapeLockdown = false
	scrapeRedactPII = false
	scrapeLocationCountry = ""
	scrapeLocationLanguages = nil
	scrapeJsonPrompt = ""
	scrapeJsonSchema = ""

	// Map
	mapSearch = ""
	mapSitemap = ""
	mapIncludeSubdomains = false
	mapIgnoreQueryParameters = false
	mapLimit = 100
	mapLocationCountry = ""
	mapLocationLanguages = nil
	mapDetailed = false

	// Search
	searchIncludeDomains = nil
	searchExcludeDomains = nil
	searchLimit = 5
	searchTBS = ""
	searchLocation = ""
	searchIgnoreInvalidURLs = false
	searchScrapeFormats = []string{"markdown"}
	searchScrapeOnlyMainContent = true

	// Agent
	agentURLs = nil
	agentSchema = ""
	agentMaxCredits = 0
	agentStrictConstrainToURLs = false
	agentModel = ""
}

// executeCommand runs a command with arguments and captures output.
func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs(args)

	err := RootCmd.Execute()
	return buf.String(), err
}

func TestScrapeCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		mockScrape    func(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error)
		wantOutput    string
		wantErrSubstr string
	}{
		{
			name: "successful markdown scrape",
			args: []string{"scrape", "https://example.com"},
			mockScrape: func(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error) {
				if url != "https://example.com" {
					return nil, errors.New("unexpected url")
				}
				if len(opts.Formats) != 1 || opts.Formats[0] != "markdown" {
					return nil, errors.New("expected default format markdown")
				}
				return &firecrawl.Document{
					Markdown: "# Example Domain\nThis is a scrape result.",
				}, nil
			},
			wantOutput: "# Example Domain\nThis is a scrape result.",
		},
		{
			name: "successful JSON output with global flag",
			args: []string{"--json", "scrape", "https://example.com"},
			mockScrape: func(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error) {
				return &firecrawl.Document{
					Markdown: "# Example",
					HTML:     "<h1>Example</h1>",
				}, nil
			},
			wantOutput: `"markdown": "# Example"`,
		},
		{
			name: "custom options mapping",
			args: []string{"scrape", "https://example.com", "--formats", "html,screenshot", "--mobile", "--wait-for", "1000", "--only-main-content=false"},
			mockScrape: func(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error) {
				if len(opts.Formats) != 2 || opts.Formats[0] != "html" || opts.Formats[1] != "screenshot" {
					return nil, fmt.Errorf("unexpected formats: %v", opts.Formats)
				}
				if opts.Mobile == nil || !*opts.Mobile {
					return nil, errors.New("expected mobile to be true")
				}
				if opts.WaitFor == nil || *opts.WaitFor != 1000 {
					return nil, errors.New("expected wait-for to be 1000")
				}
				if opts.OnlyMainContent == nil || *opts.OnlyMainContent {
					return nil, errors.New("expected only-main-content to be false")
				}
				return &firecrawl.Document{
					HTML: "<h1>Scraped HTML</h1>",
				}, nil
			},
			wantOutput: "<h1>Scraped HTML</h1>",
		},
		{
			name: "scrape error propagation",
			args: []string{"scrape", "https://invalid-domain.com"},
			mockScrape: func(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error) {
				return nil, errors.New("API error: rate limited")
			},
			wantErrSubstr: "scraping failed: API error: rate limited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			createClient = func(apiKey, apiURL string, timeout time.Duration) (FirecrawlClient, error) {
				return &mockFirecrawlClient{scrapeFn: tt.mockScrape}, nil
			}

			out, err := executeCommand(t, tt.args...)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(out, tt.wantOutput) {
					t.Fatalf("expected output containing %q, got %q", tt.wantOutput, out)
				}
			}
		})
	}
}

func TestMapCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		mockMap       func(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error)
		wantOutput    string
		wantErrSubstr string
	}{
		{
			name: "successful default map",
			args: []string{"map", "https://example.com"},
			mockMap: func(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error) {
				return &firecrawl.MapData{
					Links: []firecrawl.LinkResult{
						{URL: "https://example.com/pricing"},
						{URL: "https://example.com/about"},
					},
				}, nil
			},
			wantOutput: "https://example.com/pricing\nhttps://example.com/about",
		},
		{
			name: "successful detailed map",
			args: []string{"map", "https://example.com", "--detailed"},
			mockMap: func(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error) {
				return &firecrawl.MapData{
					Links: []firecrawl.LinkResult{
						{URL: "https://example.com/pricing", Title: "Pricing", Description: "Plan details"},
					},
				}, nil
			},
			wantOutput: "https://example.com/pricing (Title: Pricing) - Plan details",
		},
		{
			name: "json output format",
			args: []string{"--json", "map", "https://example.com"},
			mockMap: func(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error) {
				return &firecrawl.MapData{
					Links: []firecrawl.LinkResult{
						{URL: "https://example.com/blog"},
					},
				}, nil
			},
			wantOutput: `"url": "https://example.com/blog"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			createClient = func(apiKey, apiURL string, timeout time.Duration) (FirecrawlClient, error) {
				return &mockFirecrawlClient{mapFn: tt.mockMap}, nil
			}

			out, err := executeCommand(t, tt.args...)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(out, tt.wantOutput) {
					t.Fatalf("expected output containing %q, got %q", tt.wantOutput, out)
				}
			}
		})
	}
}

func TestSearchCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		mockSearch    func(ctx context.Context, query string, opts *firecrawl.SearchOptions) (*firecrawl.SearchData, error)
		wantOutput    string
		wantErrSubstr string
	}{
		{
			name: "successful web search",
			args: []string{"search", "firecrawl scraper"},
			mockSearch: func(ctx context.Context, query string, opts *firecrawl.SearchOptions) (*firecrawl.SearchData, error) {
				if query != "firecrawl scraper" {
					return nil, errors.New("unexpected query")
				}
				return &firecrawl.SearchData{
					Web: []map[string]interface{}{
						{
							"title":       "Firecrawl Github",
							"url":         "https://github.com/firecrawl",
							"description": "Clean scraper repository",
						},
					},
				}, nil
			},
			wantOutput: "=== Web Results ===\n\n[1] Firecrawl Github\n    URL: https://github.com/firecrawl\n    Snippet: Clean scraper repository",
		},
		{
			name: "search with customized scrape formats and limit",
			args: []string{"search", "test", "--limit", "3", "--scrape-formats", "markdown,html"},
			mockSearch: func(ctx context.Context, query string, opts *firecrawl.SearchOptions) (*firecrawl.SearchData, error) {
				if opts.Limit == nil || *opts.Limit != 3 {
					return nil, errors.New("expected limit of 3")
				}
				if opts.ScrapeOptions == nil || len(opts.ScrapeOptions.Formats) != 2 {
					return nil, errors.New("expected custom scrape options formats")
				}
				return &firecrawl.SearchData{
					Web: []map[string]interface{}{
						{
							"title":    "Scraped Result",
							"url":      "https://example.com/scraped",
							"markdown": "Scraped body content here",
						},
					},
				}, nil
			},
			wantOutput: "Markdown Length: 25 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			createClient = func(apiKey, apiURL string, timeout time.Duration) (FirecrawlClient, error) {
				return &mockFirecrawlClient{searchFn: tt.mockSearch}, nil
			}

			out, err := executeCommand(t, tt.args...)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(out, tt.wantOutput) {
					t.Fatalf("expected output containing %q, got %q", tt.wantOutput, out)
				}
			}
		})
	}
}

func TestAgentCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		mockAgent     func(ctx context.Context, opts *firecrawl.AgentOptions) (*firecrawl.AgentStatusResponse, error)
		wantOutput    string
		wantErrSubstr string
	}{
		{
			name: "successful agent run",
			args: []string{"agent", "Extract details", "--urls", "https://example.com"},
			mockAgent: func(ctx context.Context, opts *firecrawl.AgentOptions) (*firecrawl.AgentStatusResponse, error) {
				if opts.Prompt != "Extract details" {
					return nil, errors.New("unexpected prompt")
				}
				if len(opts.URLs) != 1 || opts.URLs[0] != "https://example.com" {
					return nil, errors.New("unexpected seed URL")
				}
				credits := 12
				return &firecrawl.AgentStatusResponse{
					Success:     true,
					Status:      "completed",
					Model:       "gpt-4o",
					CreditsUsed: &credits,
					Data: map[string]interface{}{
						"items": []string{"widget-a", "widget-b"},
					},
				}, nil
			},
			wantOutput: "Status:       completed\nSuccess:      true\nModel Used:   gpt-4o\nCredits Used: 12",
		},
		{
			name: "agent schema options string parse",
			args: []string{"agent", "Schema test", "--schema", `{"type":"object"}`},
			mockAgent: func(ctx context.Context, opts *firecrawl.AgentOptions) (*firecrawl.AgentStatusResponse, error) {
				if opts.Schema == nil || opts.Schema["type"] != "object" {
					return nil, errors.New("failed to parse schema from raw string")
				}
				return &firecrawl.AgentStatusResponse{
					Success: true,
					Status:  "completed",
				}, nil
			},
			wantOutput: "Success:      true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			createClient = func(apiKey, apiURL string, timeout time.Duration) (FirecrawlClient, error) {
				return &mockFirecrawlClient{agentFn: tt.mockAgent}, nil
			}

			out, err := executeCommand(t, tt.args...)
			if tt.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(out, tt.wantOutput) {
					t.Fatalf("expected output containing %q, got %q", tt.wantOutput, out)
				}
			}
		})
	}
}
