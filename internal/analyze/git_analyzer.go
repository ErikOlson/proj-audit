package analyze

import "github.com/your-username/proj-audit/internal/model"

type GitAnalyzer struct{}

func NewGitAnalyzer() *GitAnalyzer {
	return &GitAnalyzer{}
}

func (g *GitAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
	// TODO: implement git-based metrics
	return model.ProjectMetrics{}, nil
}
