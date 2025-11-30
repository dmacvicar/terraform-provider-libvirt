package docgen

import (
	"sort"
	"strings"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docindex"
)

// enrichReferenceHints attaches docindex-based reference hints to every field in the batch.
func enrichReferenceHints(batch Batch, index docindex.Index, limit int) Batch {
	enriched := Batch{Fields: make([]FieldContext, len(batch.Fields))}

	for i, field := range batch.Fields {
		field.ReferenceHints = matchReferenceHints(field, index, limit)
		enriched.Fields[i] = field
	}

	return enriched
}

func matchReferenceHints(field FieldContext, index docindex.Index, limit int) []ReferenceHint {
	tokens := xmlPathTokens(field.XMLPath)
	if len(tokens) == 0 {
		return nil
	}

	tokenSet := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		tokenSet[t] = struct{}{}
	}
	primary := tokens[len(tokens)-1]

	var candidates []ReferenceHint
	for _, fileIndex := range index {
		for _, section := range fileIndex.Sections {
			score := scoreSection(section, tokenSet, primary)
			if score == 0 {
				continue
			}
			candidates = append(candidates, ReferenceHint{
				Title: section.Title,
				URL:   section.URL,
				Score: score,
			})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Title < candidates[j].Title
		}
		return candidates[i].Score > candidates[j].Score
	})

	unique := make([]ReferenceHint, 0, limit)
	seen := make(map[string]bool)
	for _, cand := range candidates {
		if seen[cand.URL] {
			continue
		}
		seen[cand.URL] = true
		unique = append(unique, cand)
		if len(unique) == limit {
			break
		}
	}

	return unique
}

func scoreSection(section docindex.Section, tokens map[string]struct{}, primary string) int {
	score := 0

	for _, kw := range section.Keywords {
		kwLower := strings.ToLower(kw)
		if _, ok := tokens[kwLower]; ok {
			score += 3
			if kwLower == primary {
				score++
			}
		}
	}

	titleLower := strings.ToLower(section.Title)
	idLower := strings.ToLower(section.ID)
	for tok := range tokens {
		if tok == "" {
			continue
		}
		if strings.Contains(titleLower, tok) {
			score += 2
		}
		if strings.Contains(idLower, tok) {
			score++
		}
	}

	if primary != "" && strings.Contains(titleLower, primary) {
		score++
	}

	return score
}

func xmlPathTokens(path string) []string {
	parts := strings.Split(path, ".")
	tokens := make([]string, 0, len(parts))
	seen := make(map[string]bool)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.TrimPrefix(part, "@")
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		if seen[lower] {
			continue
		}
		seen[lower] = true
		tokens = append(tokens, lower)
	}

	return tokens
}
