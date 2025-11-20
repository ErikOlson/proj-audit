package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ErikOlson/proj-audit/internal/analyze"
	"github.com/ErikOlson/proj-audit/internal/config"
	"github.com/ErikOlson/proj-audit/internal/render"
	"github.com/ErikOlson/proj-audit/internal/scan"
	"github.com/ErikOlson/proj-audit/internal/score"
)

func main() {
	rootFlag := flag.String("root", "", "root directory to scan (default: config or current directory)")
	maxDepthFlag := flag.Int("max-depth", -1, "maximum directory depth to scan (-1 = use config, 0 = unlimited)")
	formatFlag := flag.String("format", "", "output format: tree|markdown|json (default from config)")
	configPath := flag.String("config", "", "path to JSON config file")
	ignoreFlag := flag.String("ignore", "", "comma-separated directories to ignore (appended to config)")
	includeHidden := flag.Bool("include-hidden", false, "include dot-prefixed directories")
	languagesFile := flag.String("languages", "", "path to a languages YAML file")
	disableAnalyzers := flag.String("disable-analyzers", "", "comma-separated analyzers to disable (git,fs,lang)")
	flag.Parse()

	cfg := config.DefaultConfig()

	if *configPath != "" {
		fileCfg, err := config.Load(*configPath)
		if err != nil {
			log.Fatalf("load config: %v", err)
		}
		cfg = config.Merge(cfg, fileCfg)
	}

	if *rootFlag != "" {
		cfg.Root = *rootFlag
	}
	if *maxDepthFlag >= 0 {
		cfg.MaxDepth = *maxDepthFlag
	}
	if *formatFlag != "" {
		cfg.Format = *formatFlag
	}
	if *ignoreFlag != "" {
		cfg.IgnoreDirs = append(cfg.IgnoreDirs, parseList(*ignoreFlag)...)
	}
	if *includeHidden {
		cfg.IncludeHidden = true
	}
	if *languagesFile != "" {
		cfg.LanguagesFile = *languagesFile
	}

	langs, err := cfg.ResolveLanguages()
	if err != nil {
		log.Fatalf("load languages: %v", err)
	}
	cfg.Languages = langs
	if cfg.Analyzers == nil {
		cfg.Analyzers = make(map[string]bool)
	}
	for _, name := range parseList(*disableAnalyzers) {
		cfg.Analyzers[strings.ToLower(name)] = false
	}
	analyzerToggles := cfg.EffectiveAnalyzers()

	ignoreDirs := cfg.AllIgnoreDirs()

	scanner := scan.NewDefaultScanner(ignoreDirs, cfg.IncludeHidden)

	tree, err := scanner.Scan(cfg.Root, cfg.MaxDepth)
	if err != nil {
		log.Fatalf("scan error: %v", err)
	}

	var analyzersList []analyze.Analyzer
	if analyzerToggles["git"] {
		analyzersList = append(analyzersList, analyze.NewGitAnalyzer())
	}
	if analyzerToggles["fs"] {
		analyzersList = append(analyzersList, analyze.NewFsAnalyzer(ignoreDirs, cfg.IncludeHidden))
	}
	if analyzerToggles["lang"] {
		analyzersList = append(analyzersList, analyze.NewLangAnalyzer(ignoreDirs, cfg.IncludeHidden, cfg.ExtensionMapping()))
	}
	if len(analyzersList) == 0 {
		log.Fatalf("no analyzers enabled; enable at least one")
	}
	analyzer := analyze.NewCompositeAnalyzer(analyzersList...)
	scorer := score.NewDefaultScorer(cfg.Scoring)

	if err := annotateTree(tree, analyzer, scorer); err != nil {
		log.Fatalf("annotate error: %v", err)
	}

	var r render.Renderer
	switch cfg.Format {
	case "tree":
		r = render.NewTreeRenderer()
	case "markdown":
		r = render.NewMarkdownRenderer()
	case "json":
		r = render.NewJSONRenderer()
	default:
		fmt.Fprintf(os.Stderr, "unknown format %q (expected tree|markdown|json)\n", cfg.Format)
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

func parseList(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}
