package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ErikOlson/proj-audit/internal/model"
)

func TestDefaultScannerDetectsProjects(t *testing.T) {
	root := t.TempDir()

	gitProject := filepath.Join(root, "git-project")
	if err := os.MkdirAll(filepath.Join(gitProject, ".git"), 0o755); err != nil {
		t.Fatalf("create git project: %v", err)
	}

	goProject := filepath.Join(root, "go-project")
	if err := os.MkdirAll(goProject, 0o755); err != nil {
		t.Fatalf("create go project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(goProject, "go.mod"), []byte("module example.com/test"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	ignoreDirs := []string{".git", "node_modules", "vendor"}
	scanner := NewDefaultScanner(ignoreDirs, false)
	tree, err := scanner.Scan(root, 0)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}

	if tree == nil {
		t.Fatalf("expected root node, got nil")
	}

	if node := findNodeByName(tree, filepath.Base(gitProject)); node == nil || node.Project == nil {
		t.Fatalf("expected git project node to be detected")
	}

	if node := findNodeByName(tree, filepath.Base(goProject)); node == nil || node.Project == nil {
		t.Fatalf("expected go.mod project node to be detected")
	}
}

func findNodeByName(node *model.Node, name string) *model.Node {
	if node == nil {
		return nil
	}
	if node.Name == name {
		return node
	}
	for _, child := range node.Children {
		if found := findNodeByName(child, name); found != nil {
			return found
		}
	}
	return nil
}
