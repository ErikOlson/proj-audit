package config

import _ "embed"

var (
	//go:embed languages.yaml
	defaultLanguagesYAML []byte

	//go:embed analyzers.yaml
	defaultAnalyzersYAML []byte

	//go:embed scoring.yaml
	defaultScoringYAML []byte
)
