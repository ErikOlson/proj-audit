package model

import "time"

type Project struct {
	Path     string         `json:"path"`
	Name     string         `json:"name"`
	Metrics  ProjectMetrics `json:"metrics"`
	Scores   ProjectScores  `json:"scores"`
	Category string         `json:"category"`
}

type ProjectMetrics struct {
	HasGit      bool      `json:"hasGit"`
	CommitCount int       `json:"commitCount"`
	ActiveDays  int       `json:"activeDays"`
	Languages   []string  `json:"languages"`
	Files       int       `json:"files"`
	LinesOfCode int       `json:"linesOfCode"`
	HasREADME   bool      `json:"hasReadme"`
	HasTests    bool      `json:"hasTests"`
	HasCI       bool      `json:"hasCi"`
	HasDocker   bool      `json:"hasDocker"`
	LastTouched time.Time `json:"lastTouched"`
}

type ProjectScores struct {
	Effort  int `json:"effort"`
	Polish  int `json:"polish"`
	Recency int `json:"recency"`
	Overall int `json:"overall"`
}

type Node struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Children []*Node  `json:"children"`
	Project  *Project `json:"project,omitempty"`
}
