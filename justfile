# Justfile for firecrawl task runner

# Default command to show available tasks
default:
    @just --list

# Run the CLI with arguments (e.g., `just run scrape https://example.com`)
run *args="": build
    go run main.go {{args}}

# Build the CLI binary
build:
	mkdir -p bin
	go build -o bin/firecrawl main.go

# Run the unit tests
test:
    go test -v ./...

# Format Go source files
fmt:
    go fmt ./...

# Run Go static analysis (vet)
vet:
    go vet ./...

# Run code formatting, vetting and tests
check: fmt vet test
