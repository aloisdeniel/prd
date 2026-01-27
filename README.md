# prd

A CLI tool to manage product requirements as local markdown documents in your repository. Designed to be both human-readable and AI-friendly.

## Install

```sh
go install github.com/aloisdeniel/prd@latest
```

## Quick start

Run `prd` with no arguments to launch the interactive TUI:

```sh
prd
```

Or use the CLI directly:

```sh
prd feat new --name "My Feature" --description "What it does"
prd feat ls
prd feat 001 us ls
```

## Document format

Requirements are stored as markdown files at `prd/<feature-id>-<feature-name>/prd.md`.

The mandatory sections are **Goals**, **User Stories**, and **Functional Requirements**. All other sections are optional.

```markdown
# Feature name

Detailed description of the feature.

## Goals

- Goal 1
- Goal 2

## User Stories

### US-1 | P1 | User story name

Description of the user story.

#### Acceptance Criteria

- [ ] Criterion 1
- [x] Completed criterion

#### Technical Considerations

Implementation notes.

## Functional Requirements

- FR-1: Requirement description

## Non-Goals

## Technical Considerations

## Analytics & Instrumentation

## Notes

### 2026-01-15

Note content.

## Risks & Mitigations

## Success Metrics

## Open Questions

1. **Question**
   Description.

---
*Document Version: 1.0*
*Last Updated: 2026-01-15*
```

### User stories

The ID must follow the format `US-X` where `X` is a sequential number starting from 1. A priority can optionally be added as `PX` (1 = highest, 5 = lowest, default 3). Acceptance criteria use markdown task list syntax (`- [ ]` / `- [x]`).

### Document footer

The version footer (`---` followed by version and date) is automatically maintained. Every mutation through `prd` increments the minor version and updates the date.

## TUI

Running `prd` with no arguments opens the interactive terminal UI. You can also launch it explicitly with `prd tui`.

| Key | Context | Action |
|---|---|---|
| `j`/`k`/arrows | Lists | Navigate |
| `Enter` | List item | Open |
| `Esc` | Non-root | Back |
| `q`/`Ctrl+C` | Any | Quit |
| `n` | Feature list | New feature |
| `n` | Feature detail | New user story |
| `d` | Feature detail | Delete user story |
| `Space` | User story detail | Toggle criterion |
| `c` | User story detail | Complete all criteria |
| `/` | Feature list | Search |
| `Tab` | Forms | Next field |

## CLI

### Features

```sh
prd feat ls                          # List all features with progress
prd feat <id>                        # Print full PRD (validates structure)
prd feat new --name "..." [--description "..."]  # Create feature, prints new ID
prd feat current                     # Get feature ID from current git branch
prd feat start <id>                  # Create and checkout feat/<dir> branch
```

### User stories

```sh
prd feat <id> us ls                  # List user stories with status
prd feat <id> us <us-id>             # Print user story markdown
prd feat <id> us next                # Print first incomplete story ID
prd feat <id> us new --name "..." \  # Create user story
  [--description "..."] \
  [--acceptance-criteria "..." --acceptance-criteria "..."] \
  [--technical-considerations "..."]
prd feat <id> us <us-id> accept ls    # List acceptance criteria
prd feat <id> us <us-id> complete    # Mark all criteria as done
prd feat <id> us <us-id> accept <n> complete  # Mark Nth criterion as done
prd feat <id> us <us-id> delete      # Remove user story
```

### Notes

```sh
prd feat <id> note ls                # List notes
prd feat <id> note add --content "..." # Add dated note
```

### AI integration

Generate a SKILL.md document that teaches AI coding assistants (like Claude Code) how to use `prd`:

```sh
prd skill              # Print to stdout
prd skill SKILL.md     # Save to file
```

## Example

<details>
<summary>Sample PRD document</summary>

```markdown
# Task Priority System

Add priority levels to tasks so users can focus on what matters most.

## Goals

- Allow assigning priority (high/medium/low) to any task
- Provide clear visual differentiation between priority levels
- Enable filtering and sorting by priority
- Default new tasks to medium priority

## User Stories

### US-1 | P1 | Add priority field to database

As a developer, I need to store task priority so it persists across sessions.

#### Acceptance Criteria

- [ ] Add priority column to tasks table: 'high' | 'medium' | 'low' (default 'medium')
- [ ] Generate and run migration successfully
- [ ] Typecheck passes

### US-2 | P1 | Display priority indicator on task cards

As a user, I want to see task priority at a glance so I know what needs attention first.

#### Acceptance Criteria

- [ ] Each task card shows colored priority badge (red=high, yellow=medium, gray=low)
- [ ] Priority visible without hovering or clicking
- [ ] Typecheck passes

### US-3 | P2 | Filter tasks by priority

As a user, I want to filter the task list to see only high-priority items when I'm focused.

#### Acceptance Criteria

- [ ] Filter dropdown with options: All | High | Medium | Low
- [ ] Filter persists in URL params
- [ ] Empty state message when no tasks match filter
- [ ] Typecheck passes

## Functional Requirements

- FR-1: Add priority field to tasks table
- FR-2: Display colored priority badge on each task card
- FR-3: Add priority filter dropdown to task list header

## Non-Goals

- No priority-based notifications or reminders
- No automatic priority assignment based on due date

## Technical Considerations

- Reuse existing badge component with color variants
- Filter state managed via URL search params

## Success Metrics

- Users can change priority in under 2 clicks
- High-priority tasks immediately visible at top of lists

---
*Document Version: 1.0*
*Last Updated: 2026-01-10*
```

</details>
