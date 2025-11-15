package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/your-username/proj-audit/internal/analyze"
	"github.com/your-username/proj-audit/internal/render"
	"github.com/your-username/proj-audit/internal/scan"
	"github.com/your-username/proj-audit/internal/score"
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
	// This will be implemented by the coding agent according to AGENT_TASKS.md.
	return nil
}
