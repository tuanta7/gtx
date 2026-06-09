# Repository Guidelines

## Project Structure & Module Organization

`gtx` is a Go CLI built with Cobra. Command entry points live in `cmd/`, with one file per command such as `cmd/prune.go`, `cmd/back.go`, and `cmd/shadow.go`. Shared implementation belongs under `internal/`: git operations in `internal/git/`, authentication in `internal/auth/`, profile storage in `internal/profile/`, and app constants in `internal/config/`. Reusable public packages live in `pkg/`, currently `pkg/netrc/`. Tests sit beside the code they cover and use the `_test.go` suffix.

## Build, Test, and Development Commands

- `make build`: builds the local CLI binary as `./gtx`.
- `go run . <command>`: runs the CLI without building, for example `go run . prune --help`.
- `go test ./...`: runs all package tests.
- `gofmt -w <files>`: formats edited Go files before committing.
- `make setup`: installs `cobra-cli` and prints its help.
- `make add COMMAND=name`: scaffolds a new Cobra command.

If sandboxed tooling cannot write to the default Go cache, run tests with `GOCACHE=/private/tmp/gtx-go-build go test ./...`.

## Coding Style & Naming Conventions

Use standard Go formatting with tabs via `gofmt`. Keep command flag variables near their command implementation and use clear prefixes, for example `prunePath` or `shadowProfile`. Prefer small functions that separate CLI prompting from `internal/` package logic. Wrap errors with context using `fmt.Errorf("failed to ...: %w", err)` when crossing a meaningful operation boundary.

## Testing Guidelines

Use Go’s `testing` package with `stretchr/testify/require` for assertions. Name tests after observable behavior, for example `TestPruneExcludesBackupGitDirectoryFromCommit`. Prefer temporary directories via `t.TempDir()` and local go-git repositories over fixtures that depend on external services. Add regression tests for destructive git operations and identity/authentication behavior.

## Commit & Pull Request Guidelines

The history uses conventional-style subjects such as `feat: add shadow command`, `fix: use go-git v5`, and `chore: update dependencies`. Keep commit subjects imperative, scoped, and concise. Pull requests should describe the user-facing behavior change, list tests run, and call out destructive git behavior, authentication changes, or config file impacts. Link related issues when available and include CLI output snippets for command UX changes.

## Security & Configuration Tips

Do not commit real tokens, local netrc secrets, or private repository URLs. GitHub authentication tokens are loaded through the auth package, and shadow profiles are stored under the user’s home directory in `.gtx/profiles.json`; keep tests isolated from real user configuration.
