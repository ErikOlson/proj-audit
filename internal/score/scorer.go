package score

import (
	"time"

	"github.com/your-username/proj-audit/internal/model"
)

type Scorer interface {
	Score(m model.ProjectMetrics) model.ProjectScores
	Categorize(scores model.ProjectScores, m model.ProjectMetrics) string
}

type DefaultScorer struct {
	Now func() time.Time
}

func NewDefaultScorer() *DefaultScorer {
	return &DefaultScorer{
		Now: time.Now,
	}
}

func (s *DefaultScorer) Score(m model.ProjectMetrics) model.ProjectScores {
	// TODO: implement scoring heuristics
	return model.ProjectScores{}
}

func (s *DefaultScorer) Categorize(scores model.ProjectScores, m model.ProjectMetrics) string {
	// TODO: implement categorization logic
	return ""
}
