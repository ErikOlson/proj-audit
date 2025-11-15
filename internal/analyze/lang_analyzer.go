package analyze

import "github.com/your-username/proj-audit/internal/model"

type LangAnalyzer struct{}

func NewLangAnalyzer() *LangAnalyzer {
	return &LangAnalyzer{}
}

func (l *LangAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
	// TODO: implement language detection metrics
	return model.ProjectMetrics{}, nil
}
