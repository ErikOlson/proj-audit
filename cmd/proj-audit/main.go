package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ErikOlson/proj-audit/internal/analyze"
	"github.com/ErikOlson/proj-audit/internal/render"
	"github.com/ErikOlson/proj-audit/internal/scan"
	"github.com/ErikOlson/proj-audit/internal/score"
)

func main() {
	root := flag.String("root", ".", "root directory to scan")
	maxDepth := flag.Int("max-depth", 0, "maximum directory depth to scan (0 = unlimited)")
	format := flag.String("format", "tree", "output format: tree|markdown|json")
	flag.Parse()

	scanner := scan.NewDefaultScanner()

	tree, err := scanner.Scan(*root, *maxDepth)
	if err != nil {
		log.Fatalf("scan error: %v", err)
	}

	// Initialize analyzers and scorer (stubs for now; to be filled per AGENT_TASKS.md)
	gitAnalyzer := analyze.NewGitAnalyzer()
	fsAnalyzer := analyze.NewFsAnalyzer()
	langAnalyzer := analyze.NewLangAnalyzer()
	analyzer := analyze.NewCompositeAnalyzer(gitAnalyzer, fsAnalyzer, langAnalyzer)
	scorer := score.NewDefaultScorer()

	if err := annotateTree(tree, analyzer, scorer); err != nil {
		log.Fatalf("annotate error: %v", err)
	}

	var r render.Renderer
	switch *format {
	case "tree":
		r = render.NewTreeRenderer()
	case "markdown":
		r = render.NewMarkdownRenderer()
	case "json":
		r = render.NewJSONRenderer()
	default:
		fmt.Fprintf(os.Stderr, "unknown format %q (expected tree|markdown|json)\n", *format)
		os.Exit(1)
	}

	if err := r.Render(tree, os.Stdout); err != nil {
		log.Fatalf("render error: %v", err)
	}
}

// annotateTree is a placeholder; the agent should move this into an appropriate package
// or keep it here if that remains simplest.
func annotateTree(root *scan.Node, analyzer analyze.Analyzer, scorer score.Scorer) error {
	if root == nil {
		return nil
	}

	var visit func(node *scan.Node) error
	visit = func(node *scan.Node) error {
		if node.Project != nil && analyzer != nil && scorer != nil {
			metrics, err := analyzer.Analyze(node.Path)
			if err != nil {
				return err
			}
			node.Project.Metrics = metrics
			scores := scorer.Score(metrics)
			node.Project.Scores = scores
			node.Project.Category = scorer.Categorize(scores, metrics)
		}
		for _, child := range node.Children {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}

	return visit(root)
}
