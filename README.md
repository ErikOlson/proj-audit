# proj-audit  

In this day and age, it's easy to whip things up and tinker on a bunch of prototypes. Sometimes a person needs help remembering which projects/directories were important.  
  
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


## Build

```bash
make build
```

This compiles the CLI to `./bin/proj-audit`. Use `make run` for the default invocation or run the binary directly as shown below.


## CLI Usage

### Example: tree output (default)

```bash
./bin/proj-audit --root ~/dev --format tree
```

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

### Other commands

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
- `--format` (string: `tree|markdown|json`)  
  Output format; defaults to whatever is set in config (tree by default).
- `--ignore` (string)  
  Comma-separated list of directories to skip (appended to config).
- `--include-hidden` (bool)  
  Include dot-prefixed directories instead of skipping them.
- `--config` (string)  
  Path to a JSON config file for advanced customization.


## Configuration

`proj-audit` loads defaults internally but can merge in a JSON config file:

```json
{
  "root": "~/dev",
  "maxDepth": 3,
  "format": "markdown",
  "ignoreDirs": ["tmp", "notes"],
  "languages": {
    "Haskell": {
      "extensions": [".hs"],
      "skipDirs": ["dist-newstyle"]
    }
  }
}
```

Key points:

- `ignoreDirs` entries are merged with the built-in list and affect the scanner and analyzers.
- `languages` lets you extend or override language detection, including extra extensions or directories to skip for that ecosystem.
- CLI flags always win over config values, so `proj-audit --format json` overrides whatever the file specifies.


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
