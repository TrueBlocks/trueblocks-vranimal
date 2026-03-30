# Project Rules

Read and follow these process documents for all planning and execution work:
- [./ai/Invoker.md](./ai/Invoker.md) — mode definitions and planning workflow
- [./ai/Rules.md](./ai/Rules.md) — ToDoList formatting, code generation, testing, and linting rules

# Aliases

- **aliases** [meta] — List all available aliases in a single table with columns: Group, Alias, Description. Sort rows by group then alias. Show the group name only on the first row of each group (blank for subsequent rows in the same group).
- **status** [git] — Run `make status` to show which repos have uncommitted or unpushed changes.
- **push** [git] — Commit and push all repos (including submodules). Equivalent to `make push` from the repo root.
- **summarize** [git] — Summarize all outstanding changes since the last commit across all repos and submodules. Show modified, new, and deleted files with a brief description of what changed. Use just the filename (not full path) and omit the git status column.
- **issues** [github] — List all open issues from GitHub via `gh issue list` in a markdown table with columns for issue number, title, and labels. Group rows by effort level (Small, Medium, Large, then unlabeled). **Epic handling:** First, identify epics (titles starting with "Epic:") and fetch their sub-issues via `gh issue view`. Show each epic as one row with its sub-issues summarized inline. Then, **exclude** any issue that is a sub-issue of an epic from the main table entirely — it must not appear twice.
- **implement <issue#>** [github] — Fetch the issue from GitHub via `gh issue view <issue#>`. Display the issue to the user and wait for questions or approval before proceeding. Once approved, read `./ai/Invoker.md` and `./ai/Rules.md`. Build `./ai/ToDoList.md` (overwrite if it exists) based on the issue. Enter Execution Mode and work through the ToDoList until all steps are complete, including `go vet ./...` and `go test ./...`. Do NOT commit or push — wait for the user to hand-test and explicitly approve.
- **design** [workflow] — Enter Design Mode. Explore, gather constraints, refine design docs. No task table edits unless stabilizing scope.
- **plan** [workflow] — Enter Planning Mode. Read `./ai/Invoker.md` and `./ai/Rules.md`, then build `./ai/ToDoList.md` (overwrite if it exists).
- **where are we** [workflow] — Display the current task and step from `./ai/ToDoList.md` and summarize what we are working on.
- **discuss <topics>** [workflow] — Enter design mode focused on discussing `<topics>`. Topics are the text after "discuss".
- **reconsider** [workflow] — Respond with open questions, comments, or concerns. List anything unresolved or risky before proceeding.

# Makefile Targets

The top-level Makefile delegates to `trueblocks-vranimal` before acting on itself. Only `trueblocks-vranimal` is a submodule (`vraniml` is a reference C++ codebase, not built).

- `make build` — Delegate to `trueblocks-vranimal`, which builds Go CLI binaries to `~/source/`.
- `make clean` — Remove built artifacts (`~/source/<binary>`) from all submodules.
- `make clobber` — Same as `clean` (no node_modules in this repo).
- `make lint` — Run `go vet ./...` and `golangci-lint run ./...` in submodules.
- `make test` — Run `go test ./...` in submodules that have a test target.
- `make add` — Stage all changes (`git add -A`) in all repos.
- `make commit` — Build, then stage and commit all repos. Pass `MSG="message"` for a custom commit message (default: "update").
- `make push` — Build, then stage, commit, and push all repos. Pass `MSG="message"` for a custom commit message.
- `make status` — Show which repos have uncommitted or unpushed changes.

Build is a prerequisite for `commit` and `push` — a failed build blocks both.

# GitHub

Use `gh` for all GitHub interactions (issues, PRs, checks, releases, etc.). Do not use the GitHub web API directly.

# Submodule: trueblocks-vranimal

A Go library and CLI toolkit for VRML97 parsing, validation, serialization, and boolean solid operations. Built with CGO (requires `CGO_CFLAGS="-I/opt/homebrew/include" CGO_LDFLAGS="-L/opt/homebrew/lib"` — handled by the Makefile).

| Binary | Description |
|---|---|
| `vrml-fmt` | VRML formatter |
| `vrml-validate` | VRML validator |
| `vrml-serialize` | VRML serializer |
| `solid-demo` | Solid geometry demo |
| `bool-demo` | Boolean operations demo |
| `bool-viz` | Boolean operations visualizer |
| `viewer` | VRML viewer (CGO, requires OpenGL) |

All binaries install to `~/source/`.

# Reference: vraniml

The original C++ VRML library. Read-only reference — not built by this repo's Makefile.
