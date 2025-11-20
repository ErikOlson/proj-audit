package score

import (
	"time"

	"github.com/ErikOlson/proj-audit/internal/config"
	"github.com/ErikOlson/proj-audit/internal/model"
)

type Scorer interface {
	Score(m model.ProjectMetrics) model.ProjectScores
	Categorize(scores model.ProjectScores, m model.ProjectMetrics) string
}

type DefaultScorer struct {
	Now    func() time.Time
	config *config.ScoringConfig
}

func NewDefaultScorer(cfg *config.ScoringConfig) *DefaultScorer {
	if cfg == nil {
		cfg = config.DefaultScoringConfig()
	}
	return &DefaultScorer{
		Now:    time.Now,
		config: cfg,
	}
}

func (s *DefaultScorer) Score(m model.ProjectMetrics) model.ProjectScores {
	var scores model.ProjectScores

	if len(s.config.Effort.Commit) > 0 {
		scores.Effort += pickRangePoints(m.CommitCount, s.config.Effort.Commit)
	}
	if len(s.config.Effort.Active) > 0 {
		scores.Effort += pickRangePoints(m.ActiveDays, s.config.Effort.Active)
	}

	if m.HasREADME {
		scores.Polish += s.config.Polish.Readme
	}
	if m.HasTests {
		scores.Polish += s.config.Polish.Tests
	}
	if m.HasCI {
		scores.Polish += s.config.Polish.CI
	}
	if m.HasDocker {
		scores.Polish += s.config.Polish.Docker
	}

	if !m.LastTouched.IsZero() && len(s.config.Recency) > 0 {
		ageDays := int(s.Now().Sub(m.LastTouched).Hours() / 24)
		scores.Recency += pickRecencyPoints(ageDays, s.config.Recency)
	}

	scores.Overall = scores.Effort + scores.Polish + scores.Recency
	return scores
}

func (s *DefaultScorer) Categorize(scores model.ProjectScores, m model.ProjectMetrics) string {
	cfg := s.config.Categories
	switch {
	case matchesRule(cfg.Experiment, scores, m):
		return "Experiment"
	case matchesRule(cfg.Prototype, scores, m):
		return "Prototype"
	case matchesRule(cfg.Archived, scores, m):
		return "Archived"
	case matchesRule(cfg.Product, scores, m):
		return "Product-ish"
	default:
		return "Serious"
	}
}

func pickRangePoints(value int, thresholds []config.RangeThreshold) int {
	for _, th := range thresholds {
		if value >= th.Min {
			return th.Points
		}
	}
	return 0
}

func pickRecencyPoints(ageDays int, thresholds []config.AgeThreshold) int {
	for _, th := range thresholds {
		if ageDays <= th.MaxDays {
			return th.Points
		}
	}
	return 0
}

func matchesRule(rule config.CategoryRule, scores model.ProjectScores, m model.ProjectMetrics) bool {
	if rule.CommitMax != nil && m.CommitCount > *rule.CommitMax {
		return false
	}
	if rule.EffortMax != nil && scores.Effort > *rule.EffortMax {
		return false
	}
	if rule.EffortMin != nil && scores.Effort < *rule.EffortMin {
		return false
	}
	if rule.PolishMax != nil && scores.Polish > *rule.PolishMax {
		return false
	}
	if rule.PolishMin != nil && scores.Polish < *rule.PolishMin {
		return false
	}
	if rule.RecencyMax != nil && scores.Recency > *rule.RecencyMax {
		return false
	}
	if rule.RecencyMin != nil && scores.Recency < *rule.RecencyMin {
		return false
	}
	return true
}
