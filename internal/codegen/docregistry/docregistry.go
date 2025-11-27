package docregistry

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"
	"gopkg.in/yaml.v3"
)

// Entry represents one documentation snippet tied to a Terraform path.
type Entry struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Reference   string `yaml:"reference"`

	source string
}

// fileEntries mirrors the YAML file format on disk.
type fileEntries struct {
	Entries []Entry `yaml:"entries"`
}

// Registry stores documentation entries keyed by libvirt/Terraform path.
type Registry struct {
	entries map[string]Entry
}

// Load reads every YAML file from dir and builds a registry.
func Load(dir string) (*Registry, error) {
	reg := &Registry{entries: map[string]Entry{}}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return reg, nil
		}
		return nil, fmt.Errorf("reading doc directory %s: %w", dir, err)
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(dir, name)
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading doc file %s: %w", path, err)
		}

		var parsed fileEntries
		if err := yaml.Unmarshal(contents, &parsed); err != nil {
			return nil, fmt.Errorf("parsing doc file %s: %w", path, err)
		}

		for idx, doc := range parsed.Entries {
			doc.Path = strings.TrimSpace(doc.Path)
			if doc.Path == "" {
				return nil, fmt.Errorf("doc file %s entry %d is missing a path", path, idx)
			}

			if _, exists := reg.entries[doc.Path]; exists {
				return nil, fmt.Errorf("duplicate documentation entry for %s (file %s)", doc.Path, path)
			}

			doc.source = path
			reg.entries[doc.Path] = doc
		}
	}

	return reg, nil
}

// Apply walks structIR rooted at rootPath (snake_case) adding doc strings when available.
func (r *Registry) Apply(rootPath string, root *generator.StructIR) {
	if r == nil || len(r.entries) == 0 || root == nil {
		return
	}
	rootPath = strings.TrimSpace(rootPath)
	if rootPath == "" {
		return
	}

	visited := map[string]bool{}
	r.applyStruct(root, rootPath, visited)
}

func (r *Registry) applyStruct(s *generator.StructIR, path string, visited map[string]bool) {
	if s == nil || path == "" {
		return
	}

	if visited[path] {
		return
	}
	visited[path] = true

	if entry, ok := r.entries[path]; ok {
		applyEntryToStruct(s, entry)
	}

	for _, field := range s.Fields {
		childPath := path + "." + field.TFName
		if entry, ok := r.entries[childPath]; ok {
			applyEntryToField(field, entry)
		}

		if field.IsNested && !field.IsCycle && field.NestedStruct != nil {
			r.applyStruct(field.NestedStruct, childPath, visited)
		}
	}
}

func applyEntryToStruct(s *generator.StructIR, entry Entry) {
	if entry.Description != "" {
		s.Description = entry.Description
		s.MarkdownDescription = entry.Description
	}
}

func applyEntryToField(f *generator.FieldIR, entry Entry) {
	if entry.Description != "" {
		f.Description = entry.Description
		// Build markdown with optional reference link
		md := entry.Description
		if entry.Reference != "" {
			md = fmt.Sprintf("%s\n\nSee: <%s>", md, entry.Reference)
		}
		f.MarkdownDescription = md
	}
}

// Paths returns the sorted list of documented paths (useful for debugging).
func (r *Registry) Paths() []string {
	if r == nil {
		return nil
	}
	paths := make([]string, 0, len(r.entries))
	for p := range r.entries {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// EntryForPath returns an entry and presence flag for a given path.
func (r *Registry) EntryForPath(path string) (Entry, bool) {
	if r == nil {
		return Entry{}, false
	}
	entry, ok := r.entries[path]
	return entry, ok
}
