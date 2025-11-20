package config

import "fmt"

func DefaultScoringConfig() *ScoringConfig {
	var sc ScoringConfig
	if err := decodeYAML(defaultScoringYAML, &sc); err != nil {
		panic(fmt.Sprintf("invalid default scoring yaml: %v", err))
	}
	return &sc
}
