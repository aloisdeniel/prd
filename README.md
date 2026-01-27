# prd

A small CLI tool to manage your product requirements as local documents in the repository. Flexible enough to be AI-friendly.

## Files

### Feature

The requirements are stored as markdown files stored as `prd/<feature-id>-<feature-name>/prd.md`.

A valid feature PRD must be structured as follows:

```markdown
# Feature name

Detailed description of the feature

## Goals

Description of the goals as markdown content.

## User Stories

### US-X | PX | User story name

Description of the user story as markdown content.

#### Acceptance Criteria 

- [ ] Acceptance criteria 1

#### Technical Considerations 

Description of technical considerations for this user story as markdown content. It might contains the link to diagrams or other resources.

## Functional Requirements

### FR-1: Functional requirement title
- FR-1.1: Sub-requirement description

## Non-Goals

Description of non-goals as markdown content.

## Technical Considerations 

Description of the technical architecture with potential sub sections and diagrams as markdown content.

## Analytics & Instrumentation 

Description of analytics and instrumentation as markdown content.

## Notes

### YYYY-MM-DD 

Note as markdown content.

## Risks & Mitigations

Description of risks and mitigations as markdown content.

## Success Metrics

Description of success metrics as markdown content.

## Open Questions

1. **Question**
   Detailed description of the question.

---
*Document Version: M.m*
*Last Updated: YYYY-MM-DD*
```

The only mandatory sections are the first three: **Goals**, **User Stories**, and **Functional Requirements**. The rest are optional.

For a User Story, the ID must follow the format `US-X` where `X` is a sequential number starting from 1. Optionally, a priority can be added as `PX` where `X` is a number from 1 (highest) to 5 (lowest). If the priority is not specified, it defaults to 3. Its acceptance criteria must be a checklist.

**Example:**


```markdown
# Task Priority System

Add priority levels to tasks so users can focus on what matters most. Tasks can be marked as high, medium, or low priority, with visual indicators and filtering to help users manage their workload effectively.

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

### US-002 | P1 | Display priority indicator on task cards

As a user, I want to see task priority at a glance so I know what needs attention first.

#### Acceptance Criteria

- [ ] Each task card shows colored priority badge (red=high, yellow=medium, gray=low)
- [ ] Priority visible without hovering or clicking
- [ ] Typecheck passes

### US-003 | P1 | Add priority selector to task edit

As a user, I want to change a task's priority when editing it.

#### Acceptance Criteria

- [ ] Priority dropdown in task edit modal
- [ ] Shows current priority as selected
- [ ] Saves immediately on selection change
- [ ] Typecheck passes

### US-004 | P2 | Filter tasks by priority

As a user, I want to filter the task list to see only high-priority items when I'm focused.

#### Acceptance Criteria

- [ ] Filter dropdown with options: All | High | Medium | Low
- [ ] Filter persists in URL params
- [ ] Empty state message when no tasks match filter
- [ ] Typecheck passes
- [ ] Verify in browser using dev-browser skill

## Functional Requirements

- FR-1: Add `priority` field to tasks table ('high' | 'medium' | 'low', default 'medium')
- FR-2: Display colored priority badge on each task card
- FR-3: Include priority selector in task edit modal
- FR-4: Add priority filter dropdown to task list header
- FR-5: Sort by priority within each status column (high to medium to low)

## Non-Goals

- No priority-based notifications or reminders
- No automatic priority assignment based on due date
- No priority inheritance for subtasks

## Technical Considerations

- Reuse existing badge component with color variants
- Filter state managed via URL search params
- Priority stored in database, not computed

## Success Metrics

- Users can change priority in under 2 clicks
- High-priority tasks immediately visible at top of lists
- No regression in task list performance

---
*Document Version: 1.1*
*Last Updated: 2026-01-10*
```

## CLI command

### Analyze a feature

Check that a feature PRD file is valid and follows the required structure. Print the markdown content of the prd if well formatted, else returns the list of issues found to the stderr.

```sh
prd feat <feature-id>
```

### Analyze a user story

Check that a feature PRD file is valid and follows the required structure. Print the markdown content of the user story if the prd document is well formatted, else returns the list of issues found to the stderr.

```sh
prd feat <feature-id> us <user-story-id> 
```

### List features

List all the features with their id, name, number of completed user stories, total number of user stories.

```sh
prd feat ls
```

### List user stories

List all the user stories for a given feature with their id, name, and status (completed or not).

```sh
prd feat <feature-id> us ls
```

### Current feature

Analyze the current git feature branch name to extract the feature id and prints the feature id if exists, else prints a message indicating that the current branch is not a feature branch to stderr.

```sh
prd feat current
```

### Start a feature

Create and checkout a new git feature branch for a given feature id.

```sh
prd feat start <feature-id>
```

### Next uncompleted user story

Find the next uncompleted user story for a given feature and prints the id if exists, else prints a message indicating that all user stories are completed to stderr.

```sh
prd feat <feature-id> us next
```

### New feature

Add a new feature PRD file to the `prd/` folder. Create a unique id for the feature based on existing features.

```sh
prd feat new --name "<feature-name>" --description "<feature-description>"
```

### New user story

Add a new user story to a feature PRD file. The new user story will be appended at the end of the User Stories section.

```sh
prd feat <feature-id> us new --name "<user-story-name>" --description "<user-story-description>" --acceptance-criteria "Acceptance criteria 1" "Acceptance criteria 2" --technical-considerations "<technical-considerations>"
```

### Delete user story

Delete a user story from a feature PRD file.

```sh
prd feat <feature-id> us <user-story-id> delete
```

### Mark acceptance criterion as completed

Mark an acceptance criterion as completed for a given user story.

```sh
prd feat <feature-id> us <user-story-id> accept <acceptance-criterion-number> complete
```

### Complete user story

Mark a user story as completed by checking all its acceptance criteria.

```sh
prd feat <feature-id> us <user-story-id> complete
```

### Add note

Add a note to the Notes section of a feature PRD file with the current date.

```sh
prd feat <feature-id> note add --content "<note-content>"
```

### List notes

List all notes in the Notes section of a feature PRD file.

```sh
prd feat <feature-id> note ls
```

## TUI

You can also use `prd tui` command to open an interactive terminal user interface to manage your product requirements.

```sh
prd tui
```
