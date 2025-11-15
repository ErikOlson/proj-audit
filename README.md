# proj-audit

`proj-audit` is a small CLI tool that scans a directory (for example your `~/dev` folder), discovers software projects, and gives each one a rough **“significance”** score based on:

- Effort (commits, time span, size)
- Polish (README / tests / CI / Docker / etc.)
- Recency (last modified / last commit)

It then renders a **directory tree view** with projects annotated inline, so you get context for where each project lives in your filesystem.

Example (desired) output:

```text
~/dev
├── event-notification-service   [Serious: 82 | Go | 137 commits | last: 2024-11]
├── wavekit-browser              [Product-ish: 76 | JS,Go | 85 commits | last: 2024-08]
├── experiments
│   ├── go-spike-1               [Experiment: 18 | Go | 4 commits]
│   └── rust-prototype           [Prototype: 24 | Rust | no git]
└── old-stuff
    ├── java-lab                 [Archived: 29 | Java | 22 commits | last: 2019-06]
    └── random-notes             [no project detected]
```

The intent is to help you triage:

- Which projects to **archive**
- Which to **resurrect**
- Which to **polish and showcase**


## High-level goals (v0)

- CLI tool written in **Go** (Go 1.22+).
- Run from any directory; default root is `.`.
- Detect directories that “look like projects” using simple heuristics:
  - `.git` directory present, **or**
  - Known manifest files (`go.mod`, `Cargo.toml`, `package.json`, etc.).
- For each detected project, compute basic metrics:
  - Git metrics (if git repo):
    - commit count
    - active span (first → last commit in days)
    - last commit timestamp
  - Filesystem metrics:
    - file count
    - rough lines-of-code (by extension)
    - languages (by extension)
    - presence of:
      - `README*`
      - tests
      - CI config
      - Docker-related files
  - Recency:
    - last modified time of files (fallback if no git)
- Compute simple scores:
  - `Effort`
  - `Polish`
  - `Recency`
  - `Overall` (combined)
- Categorize each project:
  - `Experiment`, `Prototype`, `Serious`, `Product-ish`, `Archived` (labels can be tuned).
- Render a **tree view** with project summary inline.
- Optionally render **JSON** or **Markdown** instead of tree.

This is intended to be extensible via **interfaces** (Scanner, Analyzer, Scorer, Renderer).


## CLI usage (desired)

Basic:

```bash
# From your dev directory
proj-audit

# From anywhere, pointing at a specific dir
proj-audit --root ~/dev

# Limit depth to keep things fast
proj-audit --root ~/dev --max-depth 3

# Output markdown
proj-audit --root ~/dev --format markdown > dev-report.md

# Output JSON (for scripting/further tooling)
proj-audit --root ~/dev --format json
```

CLI flags (v0):

- `--root` (string, default: `.`)  
  Root directory to scan.
- `--max-depth` (int, default: 0 = unlimited)  
  Maximum directory depth to recurse.
- `--format` (string: `tree|markdown|json`, default: `tree`)  
  Output format.


## Architecture

The design is intentionally interface-driven for composability and testability.

Top-level flow:

1. **Scan** filesystem → build a tree of directories.
2. For each directory that looks like a project:
   - Run **Analyzers** → produce `ProjectMetrics`.
   - Run **Scorer** → produce `ProjectScores` + category.
3. **Annotate** the tree with project info.
4. **Render** using the chosen output format (tree/md/json).


### Core domain types (concept)

```go
// Project captures all metadata for a detected project.
type Project struct {
    Path     string
    Name     string
    Metrics  ProjectMetrics
    Scores   ProjectScores
    Category string // e.g. "Experiment", "Serious", etc.
}

// ProjectMetrics are raw facts derived from analyzers.
type ProjectMetrics struct {
    HasGit      bool
    CommitCount int
    ActiveDays  int // days between first and last commit

    Languages   []string
    Files       int
    LinesOfCode int

    HasREADME   bool
    HasTests    bool
    HasCI       bool
    HasDocker   bool

    LastTouched time.Time // last commit or last modified time fallback
}

// ProjectScores are derived from metrics.
type ProjectScores struct {
    Effort  int
    Polish  int
    Recency int
    Overall int
}
```

Tree structure for directories:

```go
// Node represents a directory node in the tree.
type Node struct {
    Name     string
    Path     string
    Children []*Node
    Project  *Project // nil if this directory is not a recognized project
}
```


### Package layout

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


## Getting started

1. Rename the module in `go.mod` to your real path, e.g.:

   ```go
   module github.com/ErikOlson/proj-audit

   go 1.22
   ```

2. Build the CLI:

   ```bash
   make build
   ```

3. Run it against a directory:

   ```bash
   ./bin/proj-audit --root ~/dev --format tree
   ```

4. Use `AGENT_TASKS.md` as guidance for a coding agent (like Codex) to flesh out the implementation.
