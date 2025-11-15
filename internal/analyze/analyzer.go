package analyze

import (
	"sort"

	"github.com/ErikOlson/proj-audit/internal/model"
)

type Analyzer interface {
	Analyze(path string) (model.ProjectMetrics, error)
}

type CompositeAnalyzer struct {
	analyzers []Analyzer
}

func NewCompositeAnalyzer(analyzers ...Analyzer) *CompositeAnalyzer {
	return &CompositeAnalyzer{analyzers: analyzers}
}

func (c *CompositeAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
	var merged model.ProjectMetrics
	for _, analyzer := range c.analyzers {
		if analyzer == nil {
			continue
		}
		metrics, err := analyzer.Analyze(path)
		if err != nil {
			return model.ProjectMetrics{}, err
		}
		merged = mergeMetrics(merged, metrics)
	}
	return merged, nil
}

func mergeMetrics(a, b model.ProjectMetrics) model.ProjectMetrics {
	result := a
	result.HasGit = result.HasGit || b.HasGit
	result.HasREADME = result.HasREADME || b.HasREADME
	result.HasTests = result.HasTests || b.HasTests
	result.HasCI = result.HasCI || b.HasCI
	result.HasDocker = result.HasDocker || b.HasDocker

	result.CommitCount = maxInt(result.CommitCount, b.CommitCount)
	result.ActiveDays = maxInt(result.ActiveDays, b.ActiveDays)
	result.Files = maxInt(result.Files, b.Files)
	result.LinesOfCode = maxInt(result.LinesOfCode, b.LinesOfCode)

	result.Languages = mergeLanguages(result.Languages, b.Languages)

	if b.LastTouched.After(result.LastTouched) {
		result.LastTouched = b.LastTouched
	}

	return result
}

func mergeLanguages(existing, incoming []string) []string {
	if len(existing) == 0 && len(incoming) == 0 {
		return nil
	}

	set := make(map[string]struct{}, len(existing)+len(incoming))
	for _, lang := range existing {
		if lang == "" {
			continue
		}
		set[lang] = struct{}{}
	}
	for _, lang := range incoming {
		if lang == "" {
			continue
		}
		set[lang] = struct{}{}
	}

	if len(set) == 0 {
		return nil
	}

	out := make([]string, 0, len(set))
	for lang := range set {
		out = append(out, lang)
	}
	sort.Strings(out)
	return out
}

func maxInt(a, b int) int {
	if b > a {
		return b
	}
	return a
}
