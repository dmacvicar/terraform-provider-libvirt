package docgen

import (
	"fmt"
	"strings"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docindex"
)

// GeneratePrompt creates a prompt for AI to generate field descriptions
// It includes the full HTML index for context and enriched field metadata.
func GeneratePrompt(batch Batch, index docindex.Index) string {
	var prompt strings.Builder

	prompt.WriteString("# Terraform Provider Documentation Generation Task\n\n")
	prompt.WriteString("You are generating field descriptions for a Terraform provider that manages libvirt resources.\n\n")

	// Provide the full documentation index as context
	prompt.WriteString("## Available Documentation Sections\n\n")
	prompt.WriteString("Below is the complete libvirt documentation index. Use this to understand what each field does:\n\n")

	for htmlFile, fileIndex := range index {
		prompt.WriteString(fmt.Sprintf("### %s\n\n", htmlFile))
		for _, section := range fileIndex.Sections {
			prompt.WriteString(fmt.Sprintf("#### %s\n", section.Title))
			if section.ID != "" {
				prompt.WriteString(fmt.Sprintf("ID: `%s`\n", section.ID))
			}
			if section.URL != "" {
				prompt.WriteString(fmt.Sprintf("URL: <%s>\n", section.URL))
			}
			if len(section.Elements) > 0 {
				prompt.WriteString(fmt.Sprintf("Keywords: %s\n", strings.Join(section.Elements, ", ")))
			}
			if section.Preview != "" {
				// Limit preview to first 500 chars to keep prompt manageable
				preview := section.Preview
				if len(preview) > 500 {
					preview = preview[:500] + "..."
				}
				prompt.WriteString(fmt.Sprintf("Content: %s\n", preview))
			}
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString("## Fields to Document\n\n")
	prompt.WriteString("Generate descriptions for the following Terraform fields. Each field includes Terraform path, XML path, optional/required/computed, presence/yes-no flags, flattening hints, and any known valid values/patterns:\n\n")

	for _, field := range batch.Fields {
		prompt.WriteString(fmt.Sprintf("- **%s** (XML: `%s`)\n", field.TFPath, field.XMLPath))
		prompt.WriteString(fmt.Sprintf("  - optional: %t, required: %t, computed: %t\n", field.Optional, field.Required, field.Computed))
		if field.PresenceBoolean {
			prompt.WriteString("  - presence_boolean: true (true emits element, false/null omits)\n")
		}
		if field.StringToBool {
			prompt.WriteString(fmt.Sprintf("  - string_to_bool: yes (true=%q, false=%q)\n", field.StringToBoolTrue, field.StringToBoolFalse))
		}
		if field.FlattenedAttr != "" {
			prompt.WriteString(fmt.Sprintf("  - flattened_attr: %s (part of value+attr flattening)\n", field.FlattenedAttr))
		}
		if len(field.ValidValues) > 0 {
			prompt.WriteString(fmt.Sprintf("  - valid_values: %s\n", strings.Join(field.ValidValues, ", ")))
		}
	}

	prompt.WriteString("\n## Requirements\n\n")
	prompt.WriteString("For each field, generate a concise description (1-2 sentences) explaining what the field configures. Include constraints (valid values/patterns), optional/required/computed behavior, and presence semantics when relevant. Provide 1–2 short examples when the value is not obvious (skip examples for simple booleans).\n\n")

	prompt.WriteString("**Style guide:**\n")
	prompt.WriteString("- Write in present tense\n")
	prompt.WriteString("- Start with what the field configures/controls/sets\n")
	prompt.WriteString("- Be technical and precise\n")
	prompt.WriteString("- Keep it brief (1-2 sentences maximum)\n")
	prompt.WriteString("- Do NOT explain XML syntax or Terraform syntax\n")
	prompt.WriteString("- Use the XML path to understand the context within the libvirt schema\n")
	prompt.WriteString("- Include constraints and valid values when known; if unknown, say that the value is user-provided\n")
	prompt.WriteString("- Note presence-only behavior or yes/no string flags when applicable\n")
	prompt.WriteString("- Provide 1–2 short examples only when helpful\n\n")

	prompt.WriteString("## Output Format\n\n")
	prompt.WriteString("Respond with valid YAML in this exact format:\n\n")
	prompt.WriteString("```yaml\n")
	prompt.WriteString("entries:\n")
	for i, field := range batch.Fields {
		if i < 3 { // Show 3 examples
			prompt.WriteString(fmt.Sprintf("  - path: %s\n", field.TFPath))
			prompt.WriteString("    description: <concise description here>\n")
			prompt.WriteString("    reference: <libvirt doc URL if clear, else \"\">\n")
		}
	}
	if len(batch.Fields) > 3 {
		prompt.WriteString("  # ... continue for all fields\n")
	}
	prompt.WriteString("```\n\n")

	prompt.WriteString("**Important:** \n")
	prompt.WriteString("- Only output the YAML. Do not include any other text before or after.\n")
	prompt.WriteString("- Use a reference URL from the index when you are confident; otherwise leave it empty.\n")
	prompt.WriteString("- Do not invent defaults or side effects. If the value is unspecified in docs, treat it as user-provided.\n")

	return prompt.String()
}

// ParseYAMLResponse extracts YAML entries from the AI response
func ParseYAMLResponse(response string) (string, error) {
	// Remove code fence markers if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```yaml")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find where "entries:" starts
	lines := strings.Split(response, "\n")
	startIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "entries:") {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		// No "entries:" found, return as-is and hope it's valid
		return response, nil
	}

	// Return from entries: onwards
	return strings.Join(lines[startIdx:], "\n"), nil
}
