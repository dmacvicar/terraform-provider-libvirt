package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docgen"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docindex"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docregistry"
)

var (
	indexFile = flag.String("index", "internal/codegen/docs/.index.json", "Input index file")
	docsDir   = flag.String("docs", "internal/codegen/docs", "Documentation directory")
	outputDir = flag.String("output", "internal/codegen/docs/prompts", "Output directory for prompts")
	batchSize = flag.Int("batch-size", 20, "Fields per batch")
	dryRun    = flag.Bool("dry-run", false, "Show plan without generating")

	// Generation mode flags
	generate   = flag.Bool("generate", false, "Generate descriptions using OpenAI API")
	apiKey     = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY env var)")
	model      = flag.String("model", "gpt-4o-mini", "OpenAI model to use")
	startBatch = flag.Int("start-batch", 1, "Start from batch number (for resumption)")
	stateFile  = flag.String("state", "internal/codegen/docs/.progress.json", "State file for tracking progress")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load existing registry to avoid overwriting documented paths
	registry, err := docregistry.Load(*docsDir)
	if err != nil {
		return fmt.Errorf("loading documentation registry: %w", err)
	}

	// Build batches from reflection (no mapping needed!)
	fmt.Printf("Building batches (size=%d)...\n", *batchSize)
	batches, err := docgen.BuildBatches(registry, *batchSize)
	if err != nil {
		return fmt.Errorf("building batches: %w", err)
	}

	totalFields := 0
	for _, batch := range batches {
		totalFields += len(batch.Fields)
	}

	fmt.Printf("  Found %d fields to document\n", totalFields)
	fmt.Printf("  Created %d batches\n", len(batches))

	if len(batches) == 0 {
		fmt.Println("\n✓ No fields found!")
		return nil
	}

	if *dryRun {
		fmt.Println("\nDry run - showing first 3 batches:")
		for i := 0; i < len(batches) && i < 3; i++ {
			fmt.Printf("\nBatch %d (%d fields):\n", i+1, len(batches[i].Fields))
			for j, field := range batches[i].Fields {
				if j < 5 {
					fmt.Printf("  - %s (XML: %s)\n", field.TFPath, field.XMLPath)
				}
			}
			if len(batches[i].Fields) > 5 {
				fmt.Printf("  ... and %d more\n", len(batches[i].Fields)-5)
			}
		}
		return nil
	}

	// Load HTML index for prompts
	fmt.Printf("Loading documentation index from %s\n", *indexFile)
	index, err := docgen.GetFullIndex(*indexFile)
	if err != nil {
		return fmt.Errorf("loading index: %w", err)
	}

	// Generation mode - use OpenAI API to generate descriptions
	if *generate {
		return runGeneration(batches, index)
	}

	// Prompt generation mode - just write prompt files
	return runPromptGeneration(batches, index)
}

func runGeneration(batches []docgen.Batch, index docindex.Index) error {
	// Get API key
	key := *apiKey
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if key == "" {
		return fmt.Errorf("API key required: use --api-key or set OPENAI_API_KEY env var")
	}

	// Create OpenAI client
	client := docgen.NewOpenAIClient(key, *model)

	fmt.Printf("\nGenerating descriptions using OpenAI (%s)...\n", *model)
	fmt.Printf("  Total batches: %d\n", len(batches))
	fmt.Printf("  Starting from batch: %d\n", *startBatch)

	// Load or create state
	state, err := docgen.LoadState(*stateFile)
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	// Override start batch if specified
	if *startBatch > 1 {
		state.LastCompletedBatch = *startBatch - 1
	}

	ctx := context.Background()

	// Process each batch
	for i := state.LastCompletedBatch; i < len(batches); i++ {
		batch := batches[i]
		batchNum := i + 1

		fmt.Printf("\n[%d/%d] Processing batch...\n", batchNum, len(batches))
		fmt.Printf("  Fields: %d\n", len(batch.Fields))

		// Generate descriptions
		response, err := client.Generate(ctx, batch, index)
		if err != nil {
			return fmt.Errorf("generating batch %d: %w", batchNum, err)
		}

		// Parse and write YAML
		yamlContent, err := docgen.ParseYAMLResponse(response)
		if err != nil {
			return fmt.Errorf("parsing response for batch %d: %w", batchNum, err)
		}

		// Extract field paths for routing to correct YAML file
		fieldPaths := make([]string, len(batch.Fields))
		for i, f := range batch.Fields {
			fieldPaths[i] = f.TFPath
		}

		if err := docgen.WriteYAMLEntries(*docsDir, yamlContent, fieldPaths); err != nil {
			return fmt.Errorf("writing YAML for batch %d: %w", batchNum, err)
		}

		// Update state
		state.LastCompletedBatch = batchNum
		state.Timestamp = time.Now().Format(time.RFC3339)
		if err := docgen.SaveState(state, *stateFile); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}

		fmt.Printf("  ✓ Completed\n")

		// Brief pause to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\n✓ All batches completed!\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Review generated YAML in %s\n", *docsDir)
	fmt.Printf("3. Run: go run ./internal/codegen\n")

	return nil
}

func runPromptGeneration(batches []docgen.Batch, index docindex.Index) error {
	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Generate prompts for each batch
	fmt.Printf("\nGenerating prompts in %s...\n", *outputDir)

	for i, batch := range batches {
		prompt := docgen.GeneratePrompt(batch, index)

		filename := filepath.Join(*outputDir, fmt.Sprintf("batch_%04d.txt", i+1))
		if err := os.WriteFile(filename, []byte(prompt), 0644); err != nil {
			return fmt.Errorf("writing prompt file: %w", err)
		}

		if (i+1)%10 == 0 {
			fmt.Printf("  Generated %d/%d prompts\n", i+1, len(batches))
		}
	}

	fmt.Printf("  Generated %d prompts\n", len(batches))

	// Create a README in the prompts directory
	readme := fmt.Sprintf(`# Documentation Generation Prompts

Generated: %s
Total batches: %d
Batch size: %d fields per batch

## How This Works

This new simplified pipeline:
1. Reflects libvirtxml structs to find all fields
2. Filters out already-documented fields
3. Batches undocumented fields
4. Provides full HTML index as context to AI
5. AI generates descriptions directly (no pre-mapping needed!)

## Automated Generation (Recommended)

To automatically generate descriptions using OpenAI:

`+"```bash"+`
export OPENAI_API_KEY=your_key_here
go run ./internal/codegen/cmd/docgen --generate
`+"```"+`

Or set the API key directly:

`+"```bash"+`
go run ./internal/codegen/cmd/docgen \
  --generate \
  --api-key YOUR_OPENAI_API_KEY \
  --model gpt-4o-mini
`+"```"+`

## Manual Mode

If you prefer to process prompts manually:

1. Open each batch file (batch_0001.txt, batch_0002.txt, etc.)
2. Copy the contents to an AI chat (Claude, ChatGPT, etc.)
3. The AI will respond with YAML entries
4. Append the YAML to the appropriate file in internal/codegen/docs/
   - domain.* paths → domain.yaml
   - network.* paths → network.yaml
   - storage_pool.* paths → storage_pool.yaml
   - storage_volume.* paths → storage_volume.yaml

## What Changed

**Old pipeline:** HTML → Index → Mapping (broken) → Prompts
**New pipeline:** HTML → Index → AI Generation (direct)

The mapping step used string matching and was fundamentally flawed.
Now AI gets the full context and determines relevance semantically.
`, time.Now().Format(time.RFC3339), len(batches), *batchSize)

	readmePath := filepath.Join(*outputDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("writing README: %w", err)
	}

	fmt.Printf("\nPrompts generated successfully!\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("Option 1 - Automated (recommended):\n")
	fmt.Printf("  export OPENAI_API_KEY=your_key_here\n")
	fmt.Printf("  go run ./internal/codegen/cmd/docgen --generate\n\n")
	fmt.Printf("Option 2 - Manual:\n")
	fmt.Printf("  1. Review prompts in %s\n", *outputDir)
	fmt.Printf("  2. Process each prompt with an AI\n")
	fmt.Printf("  3. Append generated YAML to files in %s\n", *docsDir)

	return nil
}
