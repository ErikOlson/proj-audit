package scan

import "github.com/your-username/proj-audit/internal/model"

type Node = model.Node

type Scanner interface {
	Scan(root string, maxDepth int) (*model.Node, error)
}

type DefaultScanner struct {
	ignoreDirs map[string]struct{}
}

func NewDefaultScanner() *DefaultScanner {
	return &DefaultScanner{
		ignoreDirs: map[string]struct{}{
			".git":         {},
			"node_modules": {},
			"vendor":       {},
			"bin":          {},
		},
	}
}

func (s *DefaultScanner) Scan(root string, maxDepth int) (*model.Node, error) {
	// TODO: implement filesystem scanning as described in AGENT_TASKS.md
	return &model.Node{
		Name: root,
		Path: root,
	}, nil
}
