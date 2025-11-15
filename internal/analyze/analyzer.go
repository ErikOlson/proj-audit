package analyze

import "github.com/your-username/proj-audit/internal/model"

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
	// TODO: merge metrics from analyzers
	return merged, nil
}

// mergeMetrics will be implemented later.
