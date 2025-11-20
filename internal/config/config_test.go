package config

import "testing"

func TestEffectiveAnalyzers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Analyzers = map[string]bool{
		"git":  false,
		"lang": true,
	}

	effective := cfg.EffectiveAnalyzers()
	if effective["git"] {
		t.Fatalf("expected git analyzer disabled")
	}
	if !effective["lang"] {
		t.Fatalf("expected lang analyzer enabled")
	}
	if !effective["fs"] {
		t.Fatalf("expected fs analyzer to remain enabled by default")
	}
}
