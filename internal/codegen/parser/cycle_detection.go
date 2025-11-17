package parser

import "strings"

// isInAnalysisStack checks if a type name is currently being analyzed (cycle detection).
func (r *LibvirtXMLReflector) isInAnalysisStack(typeName string) bool {
	for _, name := range r.analysisStack {
		if name == typeName {
			return true
		}
	}
	return false
}

// buildCyclePath builds a human-readable cycle path for error messages.
// e.g., "DomainDiskSource → DomainDiskDataStore → DomainDiskSource"
func (r *LibvirtXMLReflector) buildCyclePath(backTo string) string {
	var path []string
	foundStart := false

	// Build path from where cycle starts
	for _, name := range r.analysisStack {
		if name == backTo {
			foundStart = true
		}
		if foundStart {
			path = append(path, name)
		}
	}

	// Add the back reference
	path = append(path, backTo)

	return strings.Join(path, " → ")
}
