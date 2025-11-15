# AGENT_TASKS.md

# Instructions for Coding Agent

You are an AI coding agent helping implement a Go CLI tool called **proj-audit**.

The goal is to scan a directory, detect software projects, compute simple metrics and scores, and render a tree with inline project summaries.

Use **Go** (Go 1.22+). Follow idiomatic Go practices. Keep dependencies to the standard library for v0.

Implement the project in phases, in this order:

1. Project skeleton (already present)
2. Model types + scanner
3. Analyzer implementations (basic)
4. Scorer
5. Renderers (tree, markdown, json)
6. Wire up `main.go`
7. Basic tests for core components


## Phase 1: Confirm project skeleton

Existing structure:

```text
proj-audit/
├── cmd/
│   └── proj-audit/
│       └── main.go
├── internal/
│   ├── model/
│   │   └── types.go
│   ├── scan/
│   │   └── scanner.go
│   ├── analyze/
│   │   ├── analyzer.go
│   │   ├── git_analyzer.go
│   │   ├── fs_analyzer.go
│   │   └── lang_analyzer.go
│   ├── score/
│   │   └── scorer.go
│   ├── render/
│   │   ├── tree_renderer.go
│   │   ├── markdown_renderer.go
│   │   └── json_renderer.go
├── go.mod
├── .gitignore
├── Makefile
├── README.md
└── AGENT_TASKS.md
```

Do **not** change this layout. Fill in the implementation inside these files.


## Phase 2: Model types + scanner

### 2.1 `internal/model/types.go`

Implement:

- `Project`
- `ProjectMetrics`
- `ProjectScores`
- `Node`

These should align with the conceptual types shown in `README.md`. Add JSON struct tags to make future JSON output stable and explicit.

Example shape:

```go
package model

import "time"

type Project struct {
    Path     string          `json:"path"`
    Name     string          `json:"name"`
    Metrics  ProjectMetrics  `json:"metrics"`
    Scores   ProjectScores   `json:"scores"`
    Category string          `json:"category"`
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
```


### 2.2 `internal/scan/scanner.go`

Define a `Scanner` interface and implement a `DefaultScanner`:

```go
package scan

import "github.com/your-username/proj-audit/internal/model"

type Scanner interface {
    Scan(root string, maxDepth int) (*model.Node, error)
}

type DefaultScanner struct {
    ignoreDirs map[string]struct{}
}

func NewDefaultScanner() *DefaultScanner {
    return &DefaultScanner{
        ignoreDirs: map[string]struct{}{
            ".git":         {},
            "node_modules": {},
            "vendor":       {},
            "bin":          {},
        },
    }
}
```

Implementation details:

- Use `filepath.WalkDir` or similar to build a tree of `model.Node` starting at `root`.
- Respect `maxDepth`:
  - `maxDepth == 0` → unlimited depth.
- A directory is a **project root** if it contains any of:
  - `.git` directory
  - `go.mod`
  - `Cargo.toml`
  - `package.json`
  - `pyproject.toml` or `requirements.txt`
  - `pom.xml` or `build.gradle`
- When a directory is a project:
  - Attach a `Project` placeholder to the node with `Name`, `Path` populated.
  - Metrics and scores will be filled later by analyzers/scorer.

Helper suggestions:

- Implement a small helper to check if a directory contains project markers.
- Use a recursive helper to build the tree rather than trying to infer hierarchy from `WalkDir` order if that’s simpler.


## Phase 3: Analyzers (basic)

### 3.1 `internal/analyze/analyzer.go`

Define:

```go
package analyze

import "github.com/your-username/proj-audit/internal/model"

type Analyzer interface {
    Analyze(path string) (model.ProjectMetrics, error)
}

type CompositeAnalyzer struct {
    analyzers []Analyzer
}

func NewCompositeAnalyzer(analyzers ...Analyzer) *CompositeAnalyzer {
    return &CompositeAnalyzer{analyzers: analyzers}
}

func (c *CompositeAnalyzer) Analyze(path string) (model.ProjectMetrics, error) {
    var merged model.ProjectMetrics
    for _, a := range c.analyzers {
        m, err := a.Analyze(path)
        if err != nil {
            return model.ProjectMetrics{}, err
        }
        merged = mergeMetrics(merged, m)
    }
    return merged, nil
}
```

Implement `mergeMetrics` sensibly:

- Boolean fields → OR them.
- Int fields → add them where aggregation makes sense, or take max (e.g., for `Files`, `LinesOfCode` if multiple analyzers contribute).
- `Languages` → merge and deduplicate.
- `LastTouched` → take the **latest** (max) between metrics.

### 3.2 `internal/analyze/git_analyzer.go`

Implement `GitAnalyzer`:

- Only applies if `.git` exists under the target path.
- Use `os.Stat` or similar to check for `.git` directory.
- Use `os/exec` to invoke `git` commands:
  - Commit count:

    ```bash
    git -C <path> rev-list --count HEAD
    ```

  - First commit timestamp (unix epoch seconds):

    ```bash
    git -C <path> log --format=%ct --reverse
    ```

    (Read first line only.)
  - Last commit timestamp:

    ```bash
    git -C <path> log -1 --format=%ct
    ```

- Convert timestamps to `time.Time` and compute `ActiveDays` as the difference between first and last in whole days.
- If git commands fail, handle gracefully and return zeroed metrics with `HasGit=false` or log/debug appropriately.

Populate:

- `HasGit`
- `CommitCount`
- `ActiveDays`
- `LastTouched` (based on last commit).


### 3.3 `internal/analyze/fs_analyzer.go`

Implement `FsAnalyzer`:

- Walk all files under the project root, skipping ignored directories:
  - `.git`, `node_modules`, `vendor`, etc.
- Count:
  - `Files` (number of non-directory entries).
  - `LinesOfCode` (simple: read file line count for text-like files; may skip very large files or known binaries).
- Detect:
  - `HasREADME` if any file named like `README`, `README.md`, `README.txt` exists at the top level.
  - `HasTests` based on simple heuristics:
    - For Go: any `_test.go` file.
    - For JS/TS: files under `__tests__/` or with suffix `.spec.js`, `.test.js`, `.spec.ts`, `.test.ts` (basic pattern matching is enough).
  - `HasCI` if any of:
    - `.github/workflows` directory exists.
    - `.gitlab-ci.yml` exists.
  - `HasDocker` if any of:
    - `Dockerfile` exists.
    - `docker-compose.yml` exists.
- Compute `LastTouched` as the latest `ModTime` among scanned files.


### 3.4 `internal/analyze/lang_analyzer.go`

Implement `LangAnalyzer`:

- Walk files under the project root.
- Infer language from file extension:
  - `.go` → `"Go"`
  - `.rs` → `"Rust"`
  - `.py` → `"Python"`
  - `.js` → `"JavaScript"`
  - `.ts` → `"TypeScript"`
  - `.java` → `"Java"`
  - `.cs` → `"C#"`
  - `.cpp`, `.cc`, `.cxx`, `.h`, `.hpp` → `"C/C++"`
- Collect languages into a set, then convert to a sorted slice for stable output.
- Populate `Languages` in `ProjectMetrics`.


## Phase 4: Scorer

### `internal/score/scorer.go`

Define:

```go
package score

import (
    "time"

    "github.com/your-username/proj-audit/internal/model"
)

type Scorer interface {
    Score(m model.ProjectMetrics) model.ProjectScores
    Categorize(scores model.ProjectScores, m model.ProjectMetrics) string
}

type DefaultScorer struct {
    Now func() time.Time
}

func NewDefaultScorer() *DefaultScorer {
    return &DefaultScorer{
        Now: time.Now,
    }
}
```

Implement `Score` using rough heuristics:

- Effort:
  - Start at 0.
  - `CommitCount`:
    - 0 → +0
    - 1–4 → +2
    - 5–19 → +5
    - 20–99 → +10
    - 100+ → +15
  - `ActiveDays`:
    - 0–1 → +0
    - 2–7 → +2
    - 8–30 → +4
    - 31+ → +6
- Polish:
  - `HasREADME` → +2
  - `HasTests` → +3
  - `HasCI` → +3
  - `HasDocker` → +2
- Recency:
  - Compute `age := Now().Sub(LastTouched)` if `LastTouched` is non-zero.
  - If `age <= 180 days` → +5
  - Else if `age <= 730 days` → +3
  - Else → +0

`Overall` can be a simple sum: `Effort + Polish + Recency`.

Implement `Categorize` roughly as:

- If `CommitCount < 5` and `Effort < 5` and `Polish < 3` → `"Experiment"`
- Else if `CommitCount < 20` and `Effort < 10` → `"Prototype"`
- Else if `Recency == 0` and `Effort >= 10` → `"Archived"`
- Else if `Polish >= 8` and `Effort >= 10` → `"Product-ish"`
- Else → `"Serious"`

The exact thresholds are not critical; clarity and usefulness are more important.


## Phase 5: Renderers

### 5.1 Tree renderer (`internal/render/tree_renderer.go`)

Implement:

```go
package render

import (
    "fmt"
    "io"
    "time"

    "github.com/your-username/proj-audit/internal/model"
)

type TreeRenderer struct{}

func NewTreeRenderer() *TreeRenderer {
    return &TreeRenderer{}
}

func (r *TreeRenderer) Render(root *model.Node, w io.Writer) error {
    return r.renderNode(root, "", true, w)
}
```

Use a recursive helper `renderNode(node *model.Node, prefix string, isLast bool, w io.Writer)` that:

- Chooses `"├── "` or `"└── "` as connector.
- Chooses `"│   "` or `"    "` as the next prefix.
- For each node:

  - Start with directory name: `prefix + connector + node.Name`
  - If `node.Project != nil`, append a summary like:

    ```text
    [Category: Overall | Lang1,Lang2 | N commits | last: YYYY-MM]
    ```

  - Derive commit count from `node.Project.Metrics.CommitCount`.
  - Derive last date from `node.Project.Metrics.LastTouched.Format("2006-01")` (if non-zero).

Render children in order, marking the last child appropriately.


### 5.2 Markdown renderer (`internal/render/markdown_renderer.go`)

Implement a `MarkdownRenderer` that:

- Flattens the tree into a slice of `*model.Project`.
- Writes:

  - `# Project Audit Report`
  - A table like:

    ```md
    | Name | Path | Category | Score | Languages | Commits | Last Touched |
    |------|------|----------|-------|-----------|---------|--------------|
    | ...  | ...  | ...      | ...   | ...       | ...     | ...          |
    ```

- Optionally, render the same tree view from `TreeRenderer` inside a code block at the bottom by reusing logic or building a small shared helper.


### 5.3 JSON renderer (`internal/render/json_renderer.go`)

Implement a `JSONRenderer` that:

- Outputs a JSON object with fields:

  ```json
  {
    "root": { ...Node... },
    "projects": [ ...flat list of Project... ]
  }
  ```

- Use `encoding/json`.
- Make sure to handle errors from `json.NewEncoder(w).Encode(...)`.


## Phase 6: Wire up `cmd/proj-audit/main.go`

In `cmd/proj-audit/main.go`:

1. Parse flags using `flag` package:
   - `root` (default `"."`)
   - `max-depth` (default `0`)
   - `format` (default `"tree"`, allowed: `"tree"`, `"markdown"`, `"json"`)
2. Construct components:
   - `scanner := scan.NewDefaultScanner()`
   - `analyzer := analyze.NewCompositeAnalyzer( /* git, fs, lang analyzers */ )`
   - `scorer := score.NewDefaultScorer()`
3. Call `tree, err := scanner.Scan(root, maxDepth)` and handle errors.
4. Walk the tree and, for each node with `node.Project != nil`:
   - `metrics, err := analyzer.Analyze(node.Path)`
   - `scores := scorer.Score(metrics)`
   - `category := scorer.Categorize(scores, metrics)`
   - Attach these to `node.Project`.
5. Select renderer based on `format`:
   - `"tree"` → `render.NewTreeRenderer()`
   - `"markdown"` → `render.NewMarkdownRenderer()`
   - `"json"` → `render.NewJSONRenderer()`
6. Call `renderer.Render(tree, os.Stdout)` and handle errors (exit with non-zero status on error).


## Phase 7: Basic tests

Add a small number of tests to validate key behavior:

- `internal/scan/scanner_test.go`:
  - Use temporary directories to verify that directories with `.git` or `go.mod` are detected as projects.
- `internal/score/scorer_test.go`:
  - Test categorization for a few representative metric combinations (Experiment, Prototype, Serious, Product-ish, Archived).

Use the standard `testing` package.


## Coding style

- Prefer small, focused functions.
- Handle errors explicitly and return them rather than panicking.
- Keep the code idiomatic and readable; performance is a non-goal for v0.
- Avoid adding external dependencies unless absolutely necessary.
