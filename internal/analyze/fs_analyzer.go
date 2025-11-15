package analyze

import "github.com/your-username/proj-audit/internal/model"

type FsAnalyzer struct{}

func NewFsAnalyzer() *FsAnalyzer {
	return &FsAnalyzer{}
}

func (f *FsAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
	// TODO: implement filesystem-based metrics
	return model.ProjectMetrics{}, nil
}
