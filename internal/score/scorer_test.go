package score

import (
	"testing"
	"time"

	"github.com/ErikOlson/proj-audit/internal/config"
	"github.com/ErikOlson/proj-audit/internal/model"
)

func TestDefaultScorerCategorize(t *testing.T) {
	now := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)

	scorer := NewDefaultScorer(config.DefaultScoringConfig())
	scorer.Now = func() time.Time { return now }

	tests := []struct {
		name     string
		metrics  model.ProjectMetrics
		category string
	}{
		{
			name: "Experiment",
			metrics: model.ProjectMetrics{
				CommitCount: 2,
				LastTouched: now,
			},
			category: "Experiment",
		},
		{
			name: "Prototype",
			metrics: model.ProjectMetrics{
				CommitCount: 10,
				ActiveDays:  3,
				LastTouched: now,
			},
			category: "Prototype",
		},
		{
			name: "Serious",
			metrics: model.ProjectMetrics{
				CommitCount: 50,
				ActiveDays:  60,
				HasREADME:   true,
				LastTouched: now,
			},
			category: "Serious",
		},
		{
			name: "Product-ish",
			metrics: model.ProjectMetrics{
				CommitCount: 150,
				ActiveDays:  200,
				HasREADME:   true,
				HasTests:    true,
				HasCI:       true,
				HasDocker:   true,
				LastTouched: now,
			},
			category: "Product-ish",
		},
		{
			name: "Archived",
			metrics: model.ProjectMetrics{
				CommitCount: 80,
				ActiveDays:  300,
				LastTouched: now.Add(-3 * 365 * 24 * time.Hour),
			},
			category: "Archived",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := scorer.Score(tt.metrics)
			got := scorer.Categorize(scores, tt.metrics)
			if got != tt.category {
				t.Fatalf("expected category %q, got %q (scores=%+v)", tt.category, got, scores)
			}
		})
	}
}
