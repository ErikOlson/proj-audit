package analyze

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type LangAnalyzer struct {
	ignoreDirs    map[string]struct{}
	includeHidden bool
	extToLang     map[string]string
}

func NewLangAnalyzer(ignoreDirs []string, includeHidden bool, extMap map[string]string) *LangAnalyzer {
	mapping := normalizeExtensionMap(extMap)
	if len(mapping) == 0 {
		mapping = defaultExtensionMap()
	}
	return &LangAnalyzer{
		ignoreDirs:    makeIgnoreSet(ignoreDirs),
		includeHidden: includeHidden,
		extToLang:     mapping,
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

		if lang := l.lookupLanguage(d.Name()); lang != "" {
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
	if !l.includeHidden && strings.HasPrefix(name, ".") && name != ".git" && name != ".github" {
		return true
	}
	_, skip := l.ignoreDirs[name]
	return skip
}

func (l *LangAnalyzer) lookupLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return ""
	}
	return l.extToLang[ext]
}

func normalizeExtensionMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for ext, lang := range in {
		ext = strings.TrimSpace(ext)
		if ext == "" || lang == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		out[strings.ToLower(ext)] = lang
	}
	return out
}

func defaultExtensionMap() map[string]string {
	return map[string]string{
		".go":   "Go",
		".rs":   "Rust",
		".py":   "Python",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".java": "Java",
		".cs":   "C#",
		".c":    "C/C++",
		".h":    "C/C++",
		".cpp":  "C/C++",
		".cc":   "C/C++",
		".cxx":  "C/C++",
		".hpp":  "C/C++",
		".hh":   "C/C++",
	}
}
