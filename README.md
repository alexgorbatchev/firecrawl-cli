# firecrawl

`firecrawl` is a Go-based command-line interface for the Firecrawl API. It enables you to scrape individual web pages, map website structures, search the web, and run autonomous, LLM-powered browser extraction agents directly from your terminal.

---

## Installation

### Prebuilt Binaries (Recommended)

To get started, download the latest prebuilt binary for your platform (Linux, macOS, or Windows) from the **GitHub Releases** page:

1. Navigate to the **Releases** tab in this GitHub repository.
2. Download the archive appropriate for your OS and CPU architecture (e.g., `firecrawl_1.0.0_linux_amd64.tar.gz`).
3. Extract the downloaded archive.
4. Move the `firecrawl` binary into a directory in your system's `PATH` (e.g., `/usr/local/bin` on macOS/Linux).

Once installed, verify that it works by running:

```bash
firecrawl --help
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

Scrapes a single page and gets its parsed content.

```bash
# Basic scrape (returns markdown content by default)
firecrawl scrape https://example.com

# Scrape and specify output formats (html and screenshots)
firecrawl scrape https://example.com --formats html,screenshot --mobile

# Perform structured JSON extraction from a web page
firecrawl scrape https://example.com --json-prompt "Extract products" --json-schema '{"type":"object"}'
```

#### Scrape Flags
- `--formats strings`: Formats to return (e.g. `markdown`, `html`, `rawHtml`, `screenshot`, `links`, `video`, `product`, `json`).
- `--only-main-content`: Only return main content of the page, excluding headers/footers (default: `true`).
- `--include-tags strings`: Comma-separated list of HTML tags to include.
- `--exclude-tags strings`: Comma-separated list of HTML tags to exclude.
- `--wait-for int`: Time in milliseconds to wait before scraping.
- `--mobile`: Scrape with mobile user-agent.
- `--skip-tls-verification`: Skip TLS certificate verification.
- `--remove-base64-images`: Strip base64-encoded images from the output.
- `--block-ads`: Block advertising scripts and elements.
- `--proxy string`: Proxy server URL (e.g., `http://proxy.example.com:8080`).
- `--max-age int`: Cache maximum age in seconds.
- `--store-in-cache`: Store scraped content in the cache.
- `--lockdown`: Enable strict lockdown browser sandbox mode.
- `--redact-pii`: Redact personally identifiable information from the scraped content.
- `--location-country string`: Geolocation targeting ISO country code (e.g., `US`, `DE`).
- `--location-languages strings`: Geolocation targeting languages (e.g., `en`, `fr`).
- `--json-prompt string`: Prompt for structured JSON extraction.
- `--json-schema string`: Raw JSON schema string or path to a JSON schema file defining the extraction structure.

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
- `--search string`: Filter mapped URLs using a search query.
- `--sitemap string`: Explicit sitemap XML URL to use for link discovery.
- `--include-subdomains`: Include subdomains in the mapped URLs.
- `--ignore-query-parameters`: Strip query strings from discovered links.
- `--limit int`: Maximum number of discovered links to return (default: `100`).
- `--detailed`: Show additional metadata (title, description) for each discovered link.
- `--location-country string`: Geolocation targeting country code.
- `--location-languages strings`: Geolocation targeting languages.

---

### 3. `search`

Searches the web and retrieves scraped results.

```bash
# Search the web for a query and output scraped markdown results
firecrawl search "firecrawl scraper"

# Limit search results and apply custom scraping rules to matched pages
firecrawl search "best coffee" --limit 3 --scrape-formats markdown
```

#### Search Flags
- `--include-domains strings`: Restrict search results to specific domains.
- `--exclude-domains strings`: Exclude specific domains from the search results.
- `--limit int`: Maximum number of search results to retrieve (default: `5`).
- `--tbs string`: Time-based search restriction (e.g. `qdr:d` for past day, `qdr:w` for past week).
- `--location string`: Location targeting parameter for Google search (e.g. `United States`).
- `--ignore-invalid-urls`: Skip scraping invalid URLs found in search results.
- `--scrape-formats strings`: Formats for scraping matched search pages (default: `markdown`).
- `--scrape-only-main-content`: Only return main content of matching pages during scrape (default: `true`).

---

### 4. `agent`

Runs an autonomous AI browser agent to crawl, discover, and extract structured data based on a prompt.

```bash
# Prompt the agent to explore and extract pricing plans
firecrawl agent "Find all plans and prices" --urls https://example.com/pricing

# Run the agent with a strict JSON schema file
firecrawl agent "Extract features" --urls https://example.com --schema ./schema.json
```

Refer to [AGENTS.md](./AGENTS.md) for a comprehensive guide on prompting, constraints, and dynamic schema configuration.

#### Agent Flags
- `--urls strings`: Seed URLs for the agent to start crawling/extracting from.
- `--schema string`: Raw JSON schema string or path to a JSON schema file defining the extracted structure.
- `--max-credits int`: Maximum credit budget permitted for the agent run.
- `--strict-constrain-to-urls`: Enforce strict agent navigation boundaries to seed URLs.
- `--model string`: The model name to use for agent reasoning and execution.

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
