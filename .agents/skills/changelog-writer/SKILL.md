---
name: changelog-writer
description: >
  Generates or updates a CHANGELOG.md in a minimal, clean Markdown format from git history,
  PR descriptions, commit messages, or a list of changes. Use this skill whenever the user wants
  to write, update, or maintain a changelog — even if they just say "update the changelog", "document
  the new release", "what changed in v2.0", or "add these changes to the changelog". Also trigger
  when the user pastes a list of changes and asks to format them, or when they provide a diff/PR
  list and want release notes. Always use this skill proactively for any changelog-related task.
---

# Changelog Writer

Produces clean, minimal changelogs in a consistent format. Inspired by [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/)
with deliberate simplifications: inline bold tags instead of sub-headings, no bracket-wrapped
versions, no link footer. These deviations are intentional — do not "correct" them.

## Output Format

Always produce Markdown in this exact structure:

```markdown
# Changelog

## v<MAJOR>.<MINOR>.<PATCH> - YYYY-MM-DD

- **<Tag>**: Description of the change, naming affected components in `backtick` code style where relevant.

## v...
```

**Date**: Include the ISO 8601 release date (`YYYY-MM-DD`) when known. Omit only if the user
hasn't provided it and it cannot be inferred — never invent a date.

**Unreleased**: If the user has changes not yet tied to a version, use `## Unreleased` as the
version heading (no date). Move it to a versioned block once a version number is provided.

**Yanked releases**: If a release was pulled due to a critical bug or security issue, append
`[YANKED]` to the version heading: `## v1.2.1 - 2024-06-01 [YANKED]`.

### Tags (use exactly these, capitalized):

| Tag               | When to use                                                         |
| ----------------- | ------------------------------------------------------------------- |
| `Added`           | New feature, function, endpoint, file, config option                |
| `Changed`         | Behavior change that is **not** breaking                            |
| `Improved`        | Enhancement to an existing feature (non-breaking)                   |
| `Deprecated`      | Feature still present but scheduled for removal                     |
| `Removed`         | Deleted feature, file, or option                                    |
| `Fixed`           | Bug fix                                                             |
| `Security`        | Vulnerability fix or security-relevant hardening                    |
| `Updated`         | Dependency bumps, tooling, CI/CD, version upgrades                  |
| `Breaking Change` | **Any** change that breaks backward compatibility — see rules below |

### Breaking Change Rules (mandatory)

Breaking changes **must** be called out explicitly. Use `**Breaking Change**` as the tag:

```markdown
- **Breaking Change**: `Config.Load()` now requires an explicit `AppName` parameter; callers
  passing `nil` will get a compile error. Update all call sites.
```

A change is breaking if it:

- Removes or renames a public function, type, method, field, or constant
- Changes the signature of a public function/method
- Changes the serialization format of persisted data
- Alters default behavior in a way that existing code/config would need updating
- Removes a previously supported environment variable or config key
- Changes a required dependency in a way that affects the public API

When in doubt, mark it as breaking and add a migration hint in the description.

## Style Rules

- One bullet per logical change. Group related sub-items only if inseparable.
- Code identifiers (function names, package names, env vars, file paths, versions) in `backtick`.
- Dependency updates: name the package and show `old → new` version in one bullet per package, or group minor bumps into one line.
- No PR numbers or author names (unless the user explicitly wants them).
- No headers other than `# Changelog` and `## v<version>`.
- No sub-headings within a version block (no `### Added`, `### Fixed` — just bullets).
- Versions in descending order (newest first).
- Be specific: "improved performance" is bad; "reduced query time for `ListUsers()` by ~40% via index on `email`" is good.
- Active voice: "Added X", not "X was added".
- Trim filler: omit "now", "also", "simply", "just", "easily".

## Input Handling

The user may give you:

1. **A list of changes in natural language** → Format them directly.
2. **Commit messages / git log** → Group and rewrite into clean bullets; skip merge commits, version bumps, and noise like "fix typo".
3. **PR titles + descriptions** → Extract the meaningful changes; ignore review chatter.
4. **A diff or code changes** → Infer what changed and write bullets accordingly.
5. **"Update the changelog" + new version info** → Prepend the new version block to the existing `# Changelog`.
6. **An existing changelog to clean up** → Reformat to match this style without losing information.

When the version number is not provided, ask for it before generating — do not invent one.

When breaking changes are present in the input (keywords: "removed", "renamed", "breaking", changed signature, dropped support), always surface them as `**Breaking Change**` bullets even if the user didn't flag them explicitly.

## Example

Input:

> v2.0.0 (released 2024-11-01): removed the old `Render()` function, added new `RenderTemplate(ctx, name, data)`,
> updated gin to 1.10, fixed nil pointer in middleware when auth header is missing,
> dropped support for Go 1.20

Output:

```markdown
## v2.0.0 - 2024-11-01

- **Breaking Change**: `Render()` has been removed. Replace all usages with `RenderTemplate(ctx, name, data)`.
- **Breaking Change**: Go 1.20 is no longer supported; minimum version is now Go 1.21.
- **Added**: `RenderTemplate(ctx, name, data)` for context-aware template rendering.
- **Fixed**: Nil pointer panic in auth middleware when the `Authorization` header is absent.
- **Updated**: `gin` (v1.9.x → v1.10).
```
