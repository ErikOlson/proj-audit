package config

func DefaultScoringConfig() *ScoringConfig {
	sc := &ScoringConfig{}
	sc.Effort.Commit = []RangeThreshold{
		{Min: 100, Points: 15},
		{Min: 20, Points: 10},
		{Min: 5, Points: 5},
		{Min: 1, Points: 2},
	}
	sc.Effort.Active = []RangeThreshold{
		{Min: 31, Points: 6},
		{Min: 8, Points: 4},
		{Min: 2, Points: 2},
	}
	sc.Polish.Readme = 2
	sc.Polish.Tests = 3
	sc.Polish.CI = 3
	sc.Polish.Docker = 2
	sc.Recency = []AgeThreshold{
		{MaxDays: 180, Points: 5},
		{MaxDays: 730, Points: 3},
	}
	sc.Categories.Experiment = CategoryRule{
		CommitMax: intPtr(4),
		EffortMax: intPtr(4),
		PolishMax: intPtr(2),
	}
	sc.Categories.Prototype = CategoryRule{
		CommitMax: intPtr(19),
		EffortMax: intPtr(9),
	}
	sc.Categories.Archived = CategoryRule{
		RecencyMax: intPtr(0),
		EffortMin:  intPtr(10),
	}
	sc.Categories.Product = CategoryRule{
		PolishMin: intPtr(8),
		EffortMin: intPtr(10),
	}
	return sc
}

func intPtr(v int) *int {
	return &v
}
