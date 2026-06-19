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
	agentMaxCredits            int
	agentStrictConstrainToURLs bool
	agentModel                 string
)

var agentCmd = &cobra.Command{
	Use:   "agent [PROMPT]",
	Short: "Run an AI-powered agent to extract structured data",
	Long:  `Run an AI-powered Firecrawl agent with a prompt to discover, navigate, and extract structured data from websites.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := args[0]

		client, err := getClient()
		if err != nil {
			return err
		}

		opts := &firecrawl.AgentOptions{
			Prompt: prompt,
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
		if agentSchema != "" {
			var schemaMap map[string]interface{}
			// Try parsing as raw JSON string first
			if err := json.Unmarshal([]byte(agentSchema), &schemaMap); err != nil {
				// If parsing fails, try reading as a file path
				fileBytes, fileErr := os.ReadFile(agentSchema)
				if fileErr != nil {
					return fmt.Errorf("schema is neither valid JSON nor a readable file path: %w (json parse err: %v)", fileErr, err)
				}
				if err := json.Unmarshal(fileBytes, &schemaMap); err != nil {
					return fmt.Errorf("parsing JSON schema from file: %w", err)
				}
			}
			opts.Schema = schemaMap
		}

		// Run the agent operation with auto-polling
		statusResp, err := client.Agent(cmd.Context(), opts)
		if err != nil {
			return fmt.Errorf("agent execution failed: %w", err)
		}

		// Output result
		if jsonOutput {
			bz, err := json.MarshalIndent(statusResp, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling status response: %w", err)
			}
			cmd.Println(string(bz))
			return nil
		}

		// Human-friendly output
		cmd.Printf("=== Agent Execution Results ===\n")
		cmd.Printf("Status:       %s\n", statusResp.Status)
		cmd.Printf("Success:      %t\n", statusResp.Success)
		if statusResp.Model != "" {
			cmd.Printf("Model Used:   %s\n", statusResp.Model)
		}
		if statusResp.CreditsUsed != nil {
			cmd.Printf("Credits Used: %d\n", *statusResp.CreditsUsed)
		}
		if statusResp.Error != "" {
			cmd.Printf("Error Details: %s\n", statusResp.Error)
		}

		if statusResp.Data != nil {
			cmd.Println("\nExtracted Data:")
			bz, err := json.MarshalIndent(statusResp.Data, "", "  ")
			if err == nil {
				cmd.Println(string(bz))
			} else {
				cmd.Printf("%+v\n", statusResp.Data)
			}
		}

		return nil
	},
}

func init() {
	// Register flags for agent command - NO shorthand single-character flags (only double-dash)
	agentCmd.Flags().StringSliceVar(&agentURLs, "urls", nil, "Seed URLs for the agent to start crawling/extracting from")
	agentCmd.Flags().StringVar(&agentSchema, "schema", "", "Raw JSON schema string or path to a JSON schema file defining extracted data structure")
	agentCmd.Flags().IntVar(&agentMaxCredits, "max-credits", 0, "Maximum credits to use for the agent execution")
	agentCmd.Flags().BoolVar(&agentStrictConstrainToURLs, "strict-constrain-to-urls", false, "Strictly restrict the agent to only visit the provided seed URLs")
	agentCmd.Flags().StringVar(&agentModel, "model", "", "The model name to use for the agent")

	RootCmd.AddCommand(agentCmd)
}
