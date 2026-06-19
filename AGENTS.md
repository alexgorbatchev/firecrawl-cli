# Firecrawl AI Agent Command Guide

The `firecrawl agent` command allows you to run autonomous, LLM-powered browser agents that navigate websites, discover links, and extract structured data based on a natural language prompt. 

This document explains how the agent works under the hood, how to construct schemas and prompts, and provides practical command examples.

---

## 1. Syntax and Core Parameters

The basic syntax for running an agent execution is:

```bash
firecrawl agent "[PROMPT]" [FLAGS]
```

### 1.1 Command-Line Flags (Double-Dash Only)

- `--urls`: (Comma-separated strings) One or more starting seed URLs where the agent should begin its crawl or analysis.
- `--schema`: (String) A raw JSON schema string or a path to a `.json` schema file on disk. This defines the structured output structure the agent will adhere to.
- `--model`: (String) Specify a custom LLM model to power the agent execution (e.g. `gpt-4o`).
- `--max-credits`: (Integer) Limit the maximum credits consumed by this single agent run to prevent accidental over-spending.
- `--strict-constrain-to-urls`: (Boolean) Restrict the agent from navigating to domains or sub-paths outside the specified seed URLs.
- `--json`: (Global Flag) Outputs the raw backend JSON response including execution metadata, rather than human-friendly summaries.

---

## 2. Dynamic JSON Schema Mapping

The `--schema` parameter is highly flexible. The CLI automatically attempts to parse your input as a raw inline JSON string first. If that parsing fails, it treats the input as a file path and reads the schema from disk.

### 2.1 Inline Schema Example
For simple structures, you can pass the JSON schema directly in the terminal:

```bash
firecrawl agent "Extract the contact email and address" \
  --urls https://example.com \
  --schema '{"type":"object","properties":{"email":{"type":"string"},"address":{"type":"string"}},"required":["email"]}'
```

### 2.2 Schema from File Example
For complex schemas, define a JSON schema file (e.g., `schema.json`):

```json
{
  "type": "object",
  "properties": {
    "pricing_plans": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "price_monthly": { "type": "string" },
          "features": {
            "type": "array",
            "items": { "type": "string" }
          }
        },
        "required": ["name", "price_monthly"]
      }
    }
  },
  "required": ["pricing_plans"]
}
```

Then invoke the CLI by referencing the file path:

```bash
firecrawl agent "Extract all monthly plans and features" \
  --urls https://example.com/pricing \
  --schema ./schema.json
```

---

## 3. Human-Friendly vs. JSON Output

By default, the CLI outputs results in an easy-to-read, structured format. When scripting or piping, pass the global `--json` flag.

### Default Output Format

```
=== Agent Execution Results ===
Status:       completed
Success:      true
Model Used:   gpt-4o
Credits Used: 14

Extracted Data:
{
  "pricing_plans": [
    {
      "features": [
        "Up to 10 users",
        "5GB storage",
        "Email support"
      ],
      "name": "Starter",
      "price_monthly": "$19"
    }
  ]
}
```

### JSON Mode Output (`--json`)

If you want to consume the complete payload in other scripts, pass the `--json` flag to return the full `AgentStatusResponse` struct:

```json
{
  "success": true,
  "status": "completed",
  "data": {
    "pricing_plans": [
      {
        "features": [
          "Up to 10 users",
          "5GB storage",
          "Email support"
        ],
        "name": "Starter",
        "price_monthly": "$19"
      }
    ]
  },
  "model": "gpt-4o",
  "expiresAt": "2026-06-18T23:59:59Z",
  "creditsUsed": 14
}
```

---

## 4. Best Practices for Prompting Agents

To get the most reliable extractions from the Firecrawl AI agent, keep the following guidelines in mind:

1. **Be Specific about Target Locations:** If you know exactly where the information resides, mention it in the prompt (e.g. "Look in the footer or the Contact Us page for social media handles").
2. **Limit Traversal Scope:** Set `--strict-constrain-to-urls` if you do not want the agent to wander onto third-party blog links, docs subdomains, or partners' sites.
3. **Use Explicit Schemas:** Standardize schemas using standard JSON types (`string`, `number`, `boolean`, `array`, `object`). Always declare `required` fields to force the agent to search harder for key data points.
4. **Constrain Credit Budgets:** High-depth agent actions can be expensive. Always use `--max-credits` to set sanity boundaries for exploratory prompt runs.
