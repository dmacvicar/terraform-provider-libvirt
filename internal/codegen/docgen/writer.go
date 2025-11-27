package docgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteYAMLEntries appends YAML entries to the appropriate documentation files
// Skips entries that already exist (deduplication)
func WriteYAMLEntries(docsDir string, yamlContent string, fields []string) error {
	// Parse YAML to extract individual entries
	entries := parseEntries(yamlContent)
	if len(entries) == 0 {
		return fmt.Errorf("no entries found in YAML content")
	}

	// Group entries by file
	fileEntries := make(map[string][]string)
	for path, entry := range entries {
		filename := getFilenameForPath(path)
		fileEntries[filename] = append(fileEntries[filename], entry)
	}

	// Append to each file
	for filename, entries := range fileEntries {
		filepath := filepath.Join(docsDir, filename)

		// Read existing file to check for existing paths
		existingPaths := make(map[string]bool)
		content, err := os.ReadFile(filepath)
		hasEntriesHeader := false
		if err == nil {
			hasEntriesHeader = strings.Contains(string(content), "entries:")
			// Extract existing paths
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "- path:") {
					path := extractPath(line)
					if path != "" {
						existingPaths[path] = true
					}
				}
			}
		}

		// Filter out duplicates
		var newEntries []string
		skipped := 0
		for _, entry := range entries {
			// Extract path from this entry
			lines := strings.Split(entry, "\n")
			var entryPath string
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "- path:") {
					entryPath = extractPath(line)
					break
				}
			}

			// Skip if already exists
			if entryPath != "" && existingPaths[entryPath] {
				skipped++
				continue
			}

			newEntries = append(newEntries, entry)
		}

		// Skip file if no new entries
		if len(newEntries) == 0 {
			if skipped > 0 {
				fmt.Printf("  Skipped %d duplicate entries in %s\n", skipped, filename)
			}
			continue
		}

		// Open file for appending
		f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening %s: %w", filepath, err)
		}
		defer func() {
			_ = f.Close()
		}()

		// If file is new or empty, add entries: header
		if !hasEntriesHeader {
			if _, err := f.WriteString("entries:\n"); err != nil {
				return fmt.Errorf("writing header to %s: %w", filepath, err)
			}
		}

		// Write new entries
		for _, entry := range newEntries {
			if _, err := f.WriteString(entry); err != nil {
				return fmt.Errorf("writing to %s: %w", filepath, err)
			}
		}

		if skipped > 0 {
			fmt.Printf("  Added %d new entries, skipped %d duplicates in %s\n", len(newEntries), skipped, filename)
		}
	}

	return nil
}

// parseEntries extracts individual YAML entries from the response
func parseEntries(yamlContent string) map[string]string {
	entries := make(map[string]string)
	lines := strings.Split(yamlContent, "\n")

	var currentPath string
	var currentEntry strings.Builder
	inEntry := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of a new entry
		if strings.HasPrefix(trimmed, "- path:") {
			// Save previous entry
			if currentPath != "" {
				entries[currentPath] = currentEntry.String()
			}

			// Start new entry
			currentPath = extractPath(trimmed)
			currentEntry.Reset()
			currentEntry.WriteString(line)
			currentEntry.WriteString("\n")
			inEntry = true
			continue
		}

		// Part of current entry
		if inEntry && (strings.HasPrefix(trimmed, "description:") || strings.HasPrefix(trimmed, "reference:")) {
			currentEntry.WriteString(line)
			currentEntry.WriteString("\n")
		}

		// End of entry (next entry or end of entries block)
		if inEntry && trimmed != "" && !strings.HasPrefix(trimmed, "description:") && !strings.HasPrefix(trimmed, "reference:") && !strings.HasPrefix(trimmed, "- path:") {
			// This might be end of entries
			if currentPath != "" {
				entries[currentPath] = currentEntry.String()
				currentPath = ""
				currentEntry.Reset()
				inEntry = false
			}
		}
	}

	// Save last entry
	if currentPath != "" {
		entries[currentPath] = currentEntry.String()
	}

	return entries
}

// extractPath extracts the path value from a "- path: xyz" line
func extractPath(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// getFilenameForPath determines which YAML file a path belongs to
func getFilenameForPath(path string) string {
	if strings.HasPrefix(path, "domain.") {
		return "domain.yaml"
	}
	if strings.HasPrefix(path, "network.") {
		return "network.yaml"
	}
	if strings.HasPrefix(path, "storage_pool.") {
		return "storage_pool.yaml"
	}
	if strings.HasPrefix(path, "storage_volume.") {
		return "storage_volume.yaml"
	}
	return "unknown.yaml"
}
