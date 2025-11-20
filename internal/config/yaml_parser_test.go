package config

import (
	"reflect"
	"testing"
)

func TestParseYAMLLanguagesLike(t *testing.T) {
	data := []byte(`
Go:
  extensions:
    - .go
  skipDirs:
    - vendor
`)

	var out map[string]LanguageConfig
	if err := decodeYAML(data, &out); err != nil {
		t.Fatalf("decodeYAML error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected one language, got %d", len(out))
	}
	if !reflect.DeepEqual(out["Go"].Extensions, []string{".go"}) {
		t.Fatalf("unexpected extensions: %#v", out["Go"].Extensions)
	}
}

func TestParseYAMLSequenceOfMaps(t *testing.T) {
	data := []byte(`
items:
  - min: 10
    points: 5
  - min: 5
    points: 2
`)

	var out map[string][]map[string]int
	if err := decodeYAML(data, &out); err != nil {
		t.Fatalf("decodeYAML error: %v", err)
	}
	if len(out["items"]) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out["items"]))
	}
	if out["items"][0]["min"] != 10 || out["items"][0]["points"] != 5 {
		t.Fatalf("unexpected first item: %+v", out["items"][0])
	}
}
