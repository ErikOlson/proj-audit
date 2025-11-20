package config

import (
	"reflect"
	"testing"
)

func TestParseLanguagesYAML(t *testing.T) {
	input := `
Go:
  extensions:
    - .go
  skipDirs:
    - vendor
Rust:
  extensions:
    - .rs
  skipDirs:
    - target
    - .cargo
"C#":
  extensions:
    - .cs
  skipDirs:
    - bin
`

	langs, err := ParseLanguagesYAML([]byte(input))
	if err != nil {
		t.Fatalf("ParseLanguagesYAML returned error: %v", err)
	}

	if len(langs) != 3 {
		t.Fatalf("expected 3 languages, got %d", len(langs))
	}

	if !reflect.DeepEqual(langs["Go"].Extensions, []string{".go"}) {
		t.Fatalf("unexpected Go extensions: %#v", langs["Go"].Extensions)
	}

	if !reflect.DeepEqual(langs["Rust"].SkipDirs, []string{"target", ".cargo"}) {
		t.Fatalf("unexpected Rust skipDirs: %#v", langs["Rust"].SkipDirs)
	}

	if _, ok := langs["C#"]; !ok {
		t.Fatalf("expected C# entry to be parsed with quotes")
	}
}

func TestParseLanguagesYAMLError(t *testing.T) {
	_, err := ParseLanguagesYAML([]byte("  extensions:\n    - .go\n"))
	if err == nil {
		t.Fatalf("expected error for field without language context")
	}
}
