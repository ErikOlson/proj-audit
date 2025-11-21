# Config Defaults

This directory contains the embedded YAML defaults that `proj-audit` ships with. They are compiled into the binary so the tool has a sensible baseline even before you provide your own config.

## Files

- `languages.yaml` – known languages, file extensions, and per-language directories to skip when scanning.
- `analyzers.yaml` – which analyzers (`git`, `fs`, `lang`) are enabled by default.
- `scoring.yaml` – the effort/polish/recency weights plus category rules that classify a project as Experiment/Prototype/Serious/etc.

## Customizing

You can override any of these on disk:

- Languages: create your own YAML (same format as `languages.yaml`) and pass `--languages path/to/file.yaml` or set `"languagesFile": "..."` in your JSON config.
- Analyzer toggles: add an `analyzers` block in `proj-audit.json` or pass `--disable-analyzers`.
- Scoring: add a `scoring` block in `proj-audit.json`. Use `scoring.yaml` here as a template.

Example: adding a Haskell language definition to your own `languages_custom.yaml`:

```yaml
Haskell:
  extensions:
    - .hs
    - .lhs
  skipDirs:
    - dist-newstyle
    - .stack-work
```

Then run:

```bash
proj-audit --languages languages_custom.yaml
```

Or reference it from your JSON config:

```json
{
  "languagesFile": "./languages_custom.yaml"
}
```

The CLI merges your file with the embedded defaults, so you only need to add new entries or override the ones you care about.
