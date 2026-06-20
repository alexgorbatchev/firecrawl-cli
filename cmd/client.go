package cmd

import (
	"context"
	"errors"
	"time"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
	"github.com/firecrawl/firecrawl/apps/go-sdk/option"
)

// FirecrawlClient defines the interface for interacting with Firecrawl.
// This interface allows mocking for tests.
type FirecrawlClient interface {
	Scrape(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error)
	Map(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error)
	Search(ctx context.Context, query string, opts *firecrawl.SearchOptions) (*firecrawl.SearchData, error)
	Agent(ctx context.Context, opts *firecrawl.AgentOptions) (*firecrawl.AgentStatusResponse, error)
	GetAgentStatus(ctx context.Context, jobID string) (*firecrawl.AgentStatusResponse, error)
	CancelAgent(ctx context.Context, jobID string) (map[string]interface{}, error)
}

// RealFirecrawlClient wraps the official SDK Client and implements FirecrawlClient.
type RealFirecrawlClient struct {
	client *firecrawl.Client
}

// NewRealFirecrawlClient creates a new instance of RealFirecrawlClient.
func NewRealFirecrawlClient(apiKey, apiURL string, timeout time.Duration) (*RealFirecrawlClient, error) {
	if apiURL == "" {
		return nil, errors.New("FIRECRAWL_API_URL or --api-url must be set.")
	}

	opts := []option.RequestOption{
		option.WithAPIURL(apiURL),
	}
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	if timeout > 0 {
		opts = append(opts, option.WithTimeout(timeout))
	}

	client, err := firecrawl.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &RealFirecrawlClient{client: client}, nil
}

// Scrape delegates to the underlying SDK Client.
func (r *RealFirecrawlClient) Scrape(ctx context.Context, url string, opts *firecrawl.ScrapeOptions) (*firecrawl.Document, error) {
	return r.client.Scrape(ctx, url, opts)
}

// Map delegates to the underlying SDK Client.
func (r *RealFirecrawlClient) Map(ctx context.Context, url string, opts *firecrawl.MapOptions) (*firecrawl.MapData, error) {
	return r.client.Map(ctx, url, opts)
}

// Search delegates to the underlying SDK Client.
func (r *RealFirecrawlClient) Search(ctx context.Context, query string, opts *firecrawl.SearchOptions) (*firecrawl.SearchData, error) {
	return r.client.Search(ctx, query, opts)
}

// Agent delegates to the underlying SDK Client's auto-polling implementation.
func (r *RealFirecrawlClient) Agent(ctx context.Context, opts *firecrawl.AgentOptions) (*firecrawl.AgentStatusResponse, error) {
	return r.client.Agent(ctx, opts)
}

// GetAgentStatus delegates to the underlying SDK Client.
func (r *RealFirecrawlClient) GetAgentStatus(ctx context.Context, jobID string) (*firecrawl.AgentStatusResponse, error) {
	return r.client.GetAgentStatus(ctx, jobID)
}

// CancelAgent delegates to the underlying SDK Client.
func (r *RealFirecrawlClient) CancelAgent(ctx context.Context, jobID string) (map[string]interface{}, error) {
	return r.client.CancelAgent(ctx, jobID)
}
