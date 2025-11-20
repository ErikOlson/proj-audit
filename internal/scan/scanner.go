package scan

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type Node = model.Node

type Scanner interface {
	Scan(root string, maxDepth int) (*model.Node, error)
}

type DefaultScanner struct {
	ignoreDirs    map[string]struct{}
	includeHidden bool
}

func NewDefaultScanner(ignoreDirs []string, includeHidden bool) *DefaultScanner {
	return &DefaultScanner{
		ignoreDirs:    toSet(ignoreDirs),
		includeHidden: includeHidden,
	}
}

func (s *DefaultScanner) Scan(root string, maxDepth int) (*model.Node, error) {
	if root == "" {
		root = "."
	}
	if maxDepth < 0 {
		maxDepth = 0
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root is not directory: %s", absRoot)
	}

	return s.scanDir(absRoot, 0, maxDepth)
}

func (s *DefaultScanner) scanDir(path string, depth, maxDepth int) (*model.Node, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", path, err)
	}

	node := &model.Node{
		Name: filepath.Base(path),
		Path: path,
	}

	if isProjectDir(path, entries) {
		node.Project = &model.Project{
			Name: node.Name,
			Path: path,
		}
	}

	if maxDepth > 0 && depth >= maxDepth {
		return node, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if s.shouldIgnore(name) {
			continue
		}
		childPath := filepath.Join(path, name)
		child, err := s.scanDir(childPath, depth+1, maxDepth)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Name < node.Children[j].Name
	})

	return node, nil
}

func (s *DefaultScanner) shouldIgnore(name string) bool {
	if name == "" {
		return false
	}
	if !s.includeHidden && strings.HasPrefix(name, ".") && name != ".git" && name != ".github" {
		return true
	}
	_, ok := s.ignoreDirs[name]
	return ok
}

func isProjectDir(path string, entries []fs.DirEntry) bool {
	markers := map[string]struct{}{
		"go.mod":           {},
		"Cargo.toml":       {},
		"package.json":     {},
		"pyproject.toml":   {},
		"requirements.txt": {},
		"pom.xml":          {},
		"build.gradle":     {},
		"build.gradle.kts": {},
	}

	hasGit := false
	hasManifest := false

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() && name == ".git" {
			hasGit = true
			break
		}
		if _, ok := markers[name]; ok {
			hasManifest = true
		}
	}

	return hasGit || hasManifest
}

func toSet(items []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	return set
}
