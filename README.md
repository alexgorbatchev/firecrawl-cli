# firecrawl

`firecrawl` is a command-line interface for the [Firecrawl](https://www.firecrawl.dev) API. It is a thin wrapper built on top of the official SDK that provides limited command-line functionality (enabling you to scrape individual web pages, map website structures, search the web, and run browser extraction agents directly from your terminal). 

Contributions and PRs are welcome to expand coverage of other SDK and API features!

---

## Installation

### Prebuilt Binaries (Recommended)

To get started, download the latest prebuilt binary for your platform (Linux, macOS, or Windows) from the [GitHub Releases](https://github.com/alexgorbatchev/firecrawl-cli/releases) page:

1. Navigate to the [Releases](https://github.com/alexgorbatchev/firecrawl-cli/releases) tab.
2. Download the archive appropriate for your OS and CPU architecture (e.g., `firecrawl_0.0.1_linux_amd64.tar.gz`).
3. Extract the downloaded archive.
4. Move the `firecrawl` binary into a directory in your system's `PATH` (e.g., `/usr/local/bin` on macOS/Linux).

Once installed, verify that it works by configuring your credentials and executing a basic scrape request targeting an imaginary internal self-hosted Firecrawl instance:

```bash
# Configure your API key and self-hosted API base URL
export FIRECRAWL_API_KEY="your-api-key"
export FIRECRAWL_API_URL="https://firecrawl.internal.co"

# Scrape a page using the direct URL positional shortcut
firecrawl https://example.com
```

---

## API Key Configuration

To interact with the Firecrawl API, you must provide your API key. The CLI reads the key from the `FIRECRAWL_API_KEY` environment variable:

```bash
export FIRECRAWL_API_KEY="your-firecrawl-api-key"
```

Alternatively, you can supply your API key with every request using the global `--api-key` flag.

---

## Global Options

The following flags are persistent across the root command and all subcommands. The CLI strictly uses **double-dash-only flags** (single-character shorthand flags are completely disabled):

- `--api-key string`: The Firecrawl API key (defaults to the `FIRECRAWL_API_KEY` environment variable).
- `--api-url string`: A custom Firecrawl base URL (defaults to the `FIRECRAWL_API_URL` environment variable).
- `--timeout duration`: The timeout for API operations (e.g., `30s`, `5m`, defaults to `5m0s`).
- `--json`: Format and output results as raw JSON instead of human-friendly terminal layouts (ideal for piping and script integrations).
- `--help`: Display the help message for `firecrawl` or any of its subcommands.

---

## Command Reference

### 1. `scrape`

Scrapes a single page and gets its parsed content. You can run it explicitly using `firecrawl scrape` or use the direct URL positional shortcut.

```bash
# Direct URL Shortcut (defaults to markdown output)
firecrawl https://example.com

# Explicit scrape command
firecrawl scrape https://example.com

# Scrape and specify output format (html and screenshots)
firecrawl scrape https://example.com --format html,screenshot --mobile

# Perform structured JSON extraction from a web page
firecrawl scrape https://example.com --format json --schema '{"type":"object","properties":{"title":{"type":"string"}}}'
```

#### Scrape Flags
- `--format strings`: Output formats (comma-separated): `markdown`, `html`, `rawHtml`, `links`, `screenshot`, `json`, `images`, `summary`, `changeTracking`, `attributes`, `branding` (default: `[markdown]`).
- `--html`: Shortcut for `--format html`.
- `--only-main-content`: Extract only main content of the page, excluding headers/footers (default: `true`).
- `--include-tags strings`: HTML tags to include (comma-separated).
- `--exclude-tags strings`: HTML tags to exclude (comma-separated).
- `--wait-for int`: Wait time in milliseconds for JS rendering.
- `--mobile`: Scrape with mobile user-agent.
- `--skip-tls-verification`: Skip TLS certificate verification.
- `--remove-base64-images`: Strip base64-encoded images from the output.
- `--block-ads`: Block advertising scripts and elements.
- `--proxy string`: Proxy mode for scraping (e.g., `auto` or `basic`).
- `--max-age int`: Cache maximum age in seconds.
- `--store-in-cache`: Store scraped content in the cache.
- `--lockdown`: Enable strict lockdown browser sandbox mode.
- `--redact-pii`: Redact personally identifiable information from the scraped content.
- `--location-country string`: Geolocation targeting ISO country code (e.g., `US`, `DE`).
- `--location-languages strings`: Geolocation targeting languages (e.g., `en`, `fr`).
- `--json-prompt string`: Prompt for structured JSON extraction.
- `--schema string`: JSON schema for structured extraction (inline JSON string).
- `--schema-file string`: Path to a JSON schema file defining the extraction structure.
- `--actions string`: JSON actions array to run during scrape (inline JSON).
- `--actions-file string`: Path to a JSON actions file.
- `--screenshot`: Take a screenshot of the page.
- `--full-page-screenshot`: Take a full page screenshot.
- `--output string`: Save output to file instead of printing.
- `--pretty`: Pretty print JSON output.
- `--timing`: Show request timing and duration information.

---

### 2. `map`

Discovers and maps URLs belonging to a website.

```bash
# Map a domain and print the list of discovered URLs
firecrawl map https://example.com

# Map domain and include subdomains
firecrawl map https://example.com --include-subdomains

# Map domain and return metadata details (titles, descriptions)
firecrawl map https://example.com --detailed --limit 50
```

#### Map Flags
- `--limit int`: Maximum URLs to discover (default: `100`).
- `--search string`: Filter discovered URLs by search query.
- `--sitemap string`: Sitemap handling mode (`include`, `skip`, `only`) or custom sitemap XML URL.
- `--include-subdomains`: Include subdomains in the mapped URLs.
- `--ignore-query-parameters`: Treat URLs with different parameters as same.
- `--wait`: Wait for map to complete.
- `--timeout int`: Timeout in seconds for the operation.
- `--detailed`: Show additional metadata (title, description) for discovered links.
- `--output string`: Save output to file instead of printing.
- `--pretty`: Pretty print JSON output.

---

### 3. `search`

Searches the web and optionally scrapes the results.

```bash
# Search the web for a query and output results
firecrawl search "firecrawl scraper"

# Limit search results and apply custom scraping rules to matched pages
firecrawl search "best coffee" --limit 3 --scrape --scrape-formats markdown
```

#### Search Flags
- `--limit int`: Maximum number of search results to retrieve (default: `5`, max: `100`).
- `--sources strings`: Sources to search (comma-separated): `web`, `images`, `news`.
- `--categories strings`: Filter by category (comma-separated): `github`, `research`, `pdf`.
- `--tbs string`: Time-based search restriction (e.g. `qdr:d` for past day, `qdr:w` for past week).
- `--location string`: Location targeting parameter for search (e.g. `Berlin,Germany`).
- `--country string`: ISO country code for geotargeting (default: `US`).
- `--timeout int`: Timeout in milliseconds (default: `60000`).
- `--ignore-invalid-urls`: Exclude URLs invalid for other Firecrawl endpoints.
- `--scrape`: Scrape search results.
- `--scrape-formats strings`: Formats for scraped content (comma-separated) (default: `[markdown]`).
- `--only-main-content`: Include only main content when scraping (default: `true`).
- `--output string`: Save output to file instead of printing.
- `--pretty`: Pretty print JSON output.

---

### 4. `agent`

Runs an autonomous AI browser agent to crawl, discover, and extract structured data based on a prompt.

```bash
# Prompt the agent to explore and extract pricing plans
firecrawl agent "Find all plans and prices" --urls https://example.com/pricing --wait

# Check status of an existing agent job using Job ID
firecrawl agent 550e8400-e29b-41d4-a716-446655440000 --status

# Cancel an active agent job using Job ID
firecrawl agent 550e8400-e29b-41d4-a716-446655440000 --cancel
```

Refer to [AGENTS.md](./AGENTS.md) for a comprehensive guide on prompting, constraints, and dynamic schema configuration.

#### Agent Flags
- `--urls strings`: Optional list of seed URLs to focus the agent on (comma-separated).
- `--model string`: Model to use for the agent run (e.g., `spark-1-mini` or `spark-1-pro`).
- `--schema string`: JSON schema for structured output (inline JSON string).
- `--schema-file string`: Path to a JSON schema file defining the extracted structure.
- `--max-credits int`: Maximum credit budget permitted for the agent run.
- `--strict-constrain-to-urls`: Enforce strict agent navigation boundaries to seed URLs.
- `--webhook string`: Webhook URL or configuration JSON.
- `--status`: Check status of an existing agent job.
- `--cancel`: Cancel an active agent job.
- `--wait`: Wait for agent to complete before returning results.
- `--poll-interval int`: Polling interval in seconds when waiting (default: `5`).
- `--timeout int`: Timeout in seconds when waiting.
- `--output string`: Save output to file instead of printing.

---

## Local Development & Compilation

To build and compile `firecrawl` locally from source:

### Requirements (Development Only)
- **Go:** 1.26 or later
- **just** (optional): Task runner for test orchestration and compilation

### Building from Source

```bash
# Using just
just build

# Or using Go directly
mkdir -p bin
go build -o bin/firecrawl main.go
```

The compiled binary is outputted to `./bin/firecrawl`.

### Code Health & Testing

The repository provides automated tasks for formatting, vetting, and unit testing:

```bash
# Run all code quality checks (formatting, vetting, testing)
just check

# Run Go static analysis (vet)
just vet

# Execute unit tests
just test

# Format all Go source files
just fmt
```

---

## Release Process

The `firecrawl` release cycle is completely automated using **GoReleaser** and GitHub Actions.

### 1. Verification
Before releasing, run all local unit tests and checks:
```bash
just check
```

### 2. Creating a Tag
Always use an annotated Git tag (`git tag -a`) to supply the release notes correctly:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
```

### 3. Pushing the Tag
Pushing the tag to GitHub triggers the release pipeline:
```bash
git push origin v1.0.0
```

The GitHub Action automatically runs verification checks, launches GoReleaser, generates prebuilt archives and a `checksums.txt` file for all targeted platforms (Darwin, Linux, Windows), and publishes them to the GitHub Releases tab of this repository.

---

## License

This project is licensed under the [MIT License](./LICENSE).
