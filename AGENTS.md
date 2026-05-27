# AGENTS.md

- Always talk in lowercase only. unless for code blocks n stuff.

---

## Hard Rules (Must Follow)

1. **NO FILE OVER 250 LINES.**
   - Break into new files instead of extending past 250 lines.
   - Group related helpers into `internal/<package>/`.
   - Use `pkg/` for code intended to be reused across multiple packages.

2. **NO COMMENTS EVER.**
   - No `//`, `/* */`, or doc comments (`// FuncName ...`).
   - If you touch a file that already has comments, do not add new ones.

3. **ONLY CHECK FOR LINT/VET ERRORS.**
   - Use `go vet ./...` as the closest equivalent to linting.

If a user request conflicts with these, ask for clarification rather than violating these rules.

---

# Caveman Mode (Agent Instructions)

## Output Rules

- Remove filler, pleasantries, and hedging.
- Drop articles and unnecessary conjunctions.
- Use short synonyms and common abbreviations.
- Use fragments when they convey meaning clearly.
- Keep technical terms exact.
- Keep code blocks unchanged.
- Quote errors exactly.

## Recommended Response Shape

`[thing] [action] [reason]. [next step].`

## Exception

Temporarily switch to normal clarity for security warnings, irreversible actions, or multi-step sequences where terseness could cause mistakes. Resume caveman mode afterward.

---

## Project Layout

- `*.go`: top-level main package or root package files
- `internal/`: private packages
- `go.mod`, `go.sum`: module definition

---

## Commands

### Install deps

```
go mod tidy
```

### "Lint" / Error Checks (Allowed)

```
go vet ./...
```

---

## Code Style

### General

- Strict idiomatic Go. no shortcuts.
- Prefer small, readable functions over clever abstractions.
- No `init()` unless absolutely necessary.

### Imports

- Use `goimports` grouping:
  1. stdlib
  2. external modules
  3. internal packages
- No unused imports (compiler enforces).

### Formatting

- `gofmt`-compliant. tabs for indentation.
- Follow existing patterns in touched files.

### Types

- Prefer explicit types on declarations when non-trivial.
- Use `error` as last return value; never ignore it.
- Prefer `T | nil` via pointers only when zero value is ambiguous.

### Naming

- Exported: `PascalCase`.
- Unexported: `camelCase`.
- Interfaces: prefer single-method names ending in `-er` (`Reader`, `Worker`).
- Booleans: `isX`, `hasX`, `shouldX`.
- Files: `snake_case.go`.

### Error Handling

- Never use `panic` in application code.
- Always propagate errors up with context: `fmt.Errorf("thing: %w", err)`.
- No blank identifier discards of errors: `_ = someCall()` is banned.
