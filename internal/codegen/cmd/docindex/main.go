package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docindex"
)

var (
	inputDir   = flag.String("input", "/usr/share/doc/libvirt/html", "Input directory containing HTML files")
	outputFile = flag.String("output", "internal/codegen/docs/.index.json", "Output JSON file")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Printf("Indexing documentation from %s\n", *inputDir)

	index := make(docindex.Index)

	// Files to process
	files := map[string]string{
		"formatdomain.html":  "https://libvirt.org/formatdomain.html",
		"formatnetwork.html": "https://libvirt.org/formatnetwork.html",
		"formatstorage.html": "https://libvirt.org/formatstorage.html",
	}

	for filename, baseURL := range files {
		path := filepath.Join(*inputDir, filename)
		fmt.Printf("  Processing %s...\n", filename)

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening %s: %w", path, err)
		}

		fileIndex, err := docindex.ParseHTML(f, baseURL)
		f.Close()
		if err != nil {
			return fmt.Errorf("parsing %s: %w", filename, err)
		}

		index[filename] = fileIndex
		fmt.Printf("    Found %d sections\n", len(fileIndex.Sections))
	}

	// Write output
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(*outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(*outputFile, data, 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	// Calculate stats
	totalSections := 0
	totalKeywords := 0
	for _, fileIdx := range index {
		totalSections += len(fileIdx.Sections)
		for _, section := range fileIdx.Sections {
			totalKeywords += len(section.Keywords)
		}
	}

	fmt.Printf("\nIndex created: %s\n", *outputFile)
	fmt.Printf("  %d sections\n", totalSections)
	fmt.Printf("  %d keywords extracted\n", totalKeywords)

	return nil
}
