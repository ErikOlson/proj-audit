package analyze

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type LangAnalyzer struct {
	ignoreDirs map[string]struct{}
}

func NewLangAnalyzer() *LangAnalyzer {
	return &LangAnalyzer{
		ignoreDirs: map[string]struct{}{
			".git":         {},
			"node_modules": {},
			"vendor":       {},
			"bin":          {},
			"target":       {},
			".gocache":     {},
			".cache":       {},
			"dist":         {},
			"build":        {},
			"out":          {},
			"venv":         {},
			".venv":        {},
			"__pycache__":  {},
			".cargo":       {},
		},
	}
}

func (l *LangAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
	languages := make(map[string]struct{})

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if p != path && l.shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if lang := inferLanguage(d.Name()); lang != "" {
			languages[lang] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return model.ProjectMetrics{}, err
	}

	if len(languages) == 0 {
		return model.ProjectMetrics{}, nil
	}

	result := model.ProjectMetrics{
		Languages: make([]string, 0, len(languages)),
	}
	for lang := range languages {
		result.Languages = append(result.Languages, lang)
	}
	sort.Strings(result.Languages)

	return result, nil
}

func (l *LangAnalyzer) shouldSkipDir(name string) bool {
	if name == "" {
		return false
	}
	if strings.HasPrefix(name, ".") && name != ".git" && name != ".github" {
		return true
	}
	_, skip := l.ignoreDirs[name]
	return skip
}

func inferLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "Go"
	case ".rs":
		return "Rust"
	case ".py":
		return "Python"
	case ".js":
		return "JavaScript"
	case ".ts":
		return "TypeScript"
	case ".java":
		return "Java"
	case ".cs":
		return "C#"
	case ".c", ".h", ".cpp", ".cc", ".cxx", ".hpp", ".hh":
		return "C/C++"
	default:
		return ""
	}
}
