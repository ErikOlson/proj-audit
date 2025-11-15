package score

import (
	"time"

	"github.com/ErikOlson/proj-audit/internal/model"
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
	var scores model.ProjectScores

	// Effort from commit count.
	switch {
	case m.CommitCount >= 100:
		scores.Effort += 15
	case m.CommitCount >= 20:
		scores.Effort += 10
	case m.CommitCount >= 5:
		scores.Effort += 5
	case m.CommitCount >= 1:
		scores.Effort += 2
	}

	// Active span contribution.
	switch {
	case m.ActiveDays >= 31:
		scores.Effort += 6
	case m.ActiveDays >= 8:
		scores.Effort += 4
	case m.ActiveDays >= 2:
		scores.Effort += 2
	}

	// Polish indicators.
	if m.HasREADME {
		scores.Polish += 2
	}
	if m.HasTests {
		scores.Polish += 3
	}
	if m.HasCI {
		scores.Polish += 3
	}
	if m.HasDocker {
		scores.Polish += 2
	}

	// Recency weighting.
	if !m.LastTouched.IsZero() {
		age := s.Now().Sub(m.LastTouched)
		if age <= 180*24*time.Hour {
			scores.Recency += 5
		} else if age <= 730*24*time.Hour {
			scores.Recency += 3
		}
	}

	scores.Overall = scores.Effort + scores.Polish + scores.Recency
	return scores
}

func (s *DefaultScorer) Categorize(scores model.ProjectScores, m model.ProjectMetrics) string {
	switch {
	case m.CommitCount < 5 && scores.Effort < 5 && scores.Polish < 3:
		return "Experiment"
	case m.CommitCount < 20 && scores.Effort < 10:
		return "Prototype"
	case scores.Recency == 0 && scores.Effort >= 10:
		return "Archived"
	case scores.Polish >= 8 && scores.Effort >= 10:
		return "Product-ish"
	default:
		return "Serious"
	}
}
