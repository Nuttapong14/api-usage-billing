package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Spec represents a loaded OpenAPI contract file.
type Spec struct {
	Name    string
	Path    string
	Content []byte
}

// LoadSpecs loads all contract files from the provided directory.
func LoadSpecs(t *testing.T, contractsDir string) []Spec {
	t.Helper()

	if contractsDir == "" {
		contractsDir = filepath.Join("..", "..", "..", "specs", "001-api-usage-billing", "contracts")
	}

	entries, err := filepath.Glob(filepath.Join(contractsDir, "*.yaml"))
	if err != nil {
		t.Fatalf("failed to list contracts in %s: %v", contractsDir, err)
	}

	if len(entries) == 0 {
		t.Fatalf("no contract files found in %s", contractsDir)
	}

	var specs []Spec
	for _, entry := range entries {
		content, err := os.ReadFile(entry)
		if err != nil {
			t.Fatalf("failed to read %s: %v", entry, err)
		}

		specs = append(specs, Spec{
			Name:    filepath.Base(entry),
			Path:    entry,
			Content: content,
		})
	}

	return specs
}

// ValidateSpecs performs lightweight checks on the loaded contracts.
func ValidateSpecs(t *testing.T, specs []Spec) {
	t.Helper()
	for _, spec := range specs {
		payload := string(spec.Content)
		if !strings.Contains(payload, "openapi:") {
			t.Fatalf("%s missing openapi header", spec.Path)
		}
		if !strings.Contains(payload, "\npaths:") {
			t.Fatalf("%s missing paths section", spec.Path)
		}
	}
}
