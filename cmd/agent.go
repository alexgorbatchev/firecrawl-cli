package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	firecrawl "github.com/firecrawl/firecrawl/apps/go-sdk"
	"github.com/spf13/cobra"
)

var (
	// Local agent flag variables
	agentURLs                  []string
	agentSchema                string
	agentSchemaFile            string
	agentMaxCredits            int
	agentStrictConstrainToURLs bool
	agentModel                 string
	agentWebhook               string
	agentStatus                bool
	agentCancel                bool
	agentWait                  bool
	agentPollInterval          int
	agentTimeout               int
	agentOutput                string
)

var agentCmd = &cobra.Command{
	Use:   "agent [PROMPT/JOB-ID]",
	Short: "Search and gather data from the web using natural language prompts",
	Long:  `Run an AI-powered Firecrawl agent with a prompt to discover, navigate, and extract structured data from websites.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		argVal := args[0]

		client, err := getClient()
		if err != nil {
			return err
		}

		// Handle check status or cancel active job
		if agentStatus || agentCancel {
			jobID := argVal
			var statusResp *firecrawl.AgentStatusResponse
			var err error

			if agentCancel {
				cancelData, cErr := client.CancelAgent(cmd.Context(), jobID)
				if cErr != nil {
					return fmt.Errorf("canceling agent job: %w", cErr)
				}
				bz, _ := json.MarshalIndent(cancelData, "", "  ")
				cmd.Println(string(bz))
				return nil
			}

			// Retrieve job status
			statusResp, err = client.GetAgentStatus(cmd.Context(), jobID)
			if err != nil {
				return fmt.Errorf("retrieving agent job status: %w", err)
			}

			return printAgentStatus(cmd, statusResp)
		}

		// Start a new agent job
		opts := &firecrawl.AgentOptions{
			Prompt: argVal,
		}

		if cmd.Flags().Changed("urls") {
			opts.URLs = agentURLs
		}
		if cmd.Flags().Changed("max-credits") {
			opts.MaxCredits = firecrawl.Int(agentMaxCredits)
		}
		if cmd.Flags().Changed("strict-constrain-to-urls") {
			opts.StrictConstrainToURLs = firecrawl.Bool(agentStrictConstrainToURLs)
		}
		if cmd.Flags().Changed("model") {
			opts.Model = firecrawl.String(agentModel)
		}

		// Handle structured extraction Schema
		if agentSchema != "" || agentSchemaFile != "" {
			var schemaMap map[string]interface{}
			var schemaBytes []byte
			var fileErr error

			if agentSchemaFile != "" {
				schemaBytes, fileErr = os.ReadFile(agentSchemaFile)
				if fileErr != nil {
					return fmt.Errorf("reading schema file: %w", fileErr)
				}
			} else {
				schemaBytes = []byte(agentSchema)
			}

			if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
				return fmt.Errorf("parsing JSON schema: %w", err)
			}
			opts.Schema = schemaMap
		}

		// Handle Webhook parameter
		if agentWebhook != "" {
			opts.Webhook = &firecrawl.WebhookConfig{}
			// Try parsing as JSON first
			if err := json.Unmarshal([]byte(agentWebhook), opts.Webhook); err != nil {
				// If not valid JSON, treat it as a raw destination URL
				opts.Webhook.URL = agentWebhook
			}
		}

		// Run the agent operation with auto-polling
		statusResp, err := client.Agent(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("agent execution failed: %w", err)
		}

		return printAgentStatus(cmd, statusResp)
	},
}

func printAgentStatus(cmd *cobra.Command, statusResp *firecrawl.AgentStatusResponse) error {
	var outputStr string

	if jsonOutput {
		bz, err := json.MarshalIndent(statusResp, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling status response: %w", err)
		}
		outputStr = string(bz)
	} else {
		outputStr = "=== Agent Execution Results ===\n"
		outputStr += fmt.Sprintf("Status:       %s\n", statusResp.Status)
		outputStr += fmt.Sprintf("Success:      %t\n", statusResp.Success)
		if statusResp.Model != "" {
			outputStr += fmt.Sprintf("Model Used:   %s\n", statusResp.Model)
		}
		if statusResp.CreditsUsed != nil {
			outputStr += fmt.Sprintf("Credits Used: %d\n", *statusResp.CreditsUsed)
		}
		if statusResp.Error != "" {
			outputStr += fmt.Sprintf("Error Details: %s\n", statusResp.Error)
		}

		if statusResp.Data != nil {
			outputStr += "\nExtracted Data:\n"
			bz, err := json.MarshalIndent(statusResp.Data, "", "  ")
			if err == nil {
				outputStr += string(bz)
			} else {
				outputStr += fmt.Sprintf("%+v", statusResp.Data)
			}
		}
	}

	// Write output to file if requested, otherwise print to stdout
	if agentOutput != "" {
		err := os.WriteFile(agentOutput, []byte(outputStr), 0644)
		if err != nil {
			return fmt.Errorf("writing output to file: %w", err)
		}
	} else {
		cmd.Println(outputStr)
	}

	return nil
}

func init() {
	// Register flags for agent command - NO shorthand single-character flags (only double-dash)
	agentCmd.Flags().StringSliceVar(&agentURLs, "urls", nil, "Optional list of URLs to focus the agent on (comma-separated)")
	agentCmd.Flags().StringVar(&agentSchema, "schema", "", "JSON schema for structured output (inline JSON string)")
	agentCmd.Flags().StringVar(&agentSchemaFile, "schema-file", "", "Path to JSON schema file for structured output")
	agentCmd.Flags().IntVar(&agentMaxCredits, "max-credits", 0, "Maximum credits to spend (job fails if limit reached)")
	agentCmd.Flags().BoolVar(&agentStrictConstrainToURLs, "strict-constrain-to-urls", false, "Strictly restrict the agent to only visit the provided seed URLs")
	agentCmd.Flags().StringVar(&agentModel, "model", "", "Model to use: spark-1-mini or spark-1-pro")
	agentCmd.Flags().StringVar(&agentWebhook, "webhook", "", "Webhook URL or configuration JSON")
	agentCmd.Flags().BoolVar(&agentStatus, "status", false, "Check status of existing agent job")
	agentCmd.Flags().BoolVar(&agentCancel, "cancel", false, "Cancel an active agent job by job ID")
	agentCmd.Flags().BoolVar(&agentWait, "wait", false, "Wait for agent to complete before returning results")
	agentCmd.Flags().IntVar(&agentPollInterval, "poll-interval", 5, "Polling interval when waiting")
	agentCmd.Flags().IntVar(&agentTimeout, "timeout", 0, "Timeout when waiting")
	agentCmd.Flags().StringVar(&agentOutput, "output", "", "Save output to file")

	RootCmd.AddCommand(agentCmd)
}
