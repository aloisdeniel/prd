package prd

import (
	"strings"
	"testing"
)

const validDocument = `# Task Priority System

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

### US-2 | P1 | Display priority indicator on task cards

As a user, I want to see task priority at a glance so I know what needs attention first.

#### Acceptance Criteria

- [ ] Each task card shows colored priority badge (red=high, yellow=medium, gray=low)
- [ ] Priority visible without hovering or clicking
- [ ] Typecheck passes

### US-3 | P1 | Add priority selector to task edit

As a user, I want to change a task's priority when editing it.

#### Acceptance Criteria

- [ ] Priority dropdown in task edit modal
- [ ] Shows current priority as selected
- [ ] Saves immediately on selection change
- [ ] Typecheck passes

### US-4 | P2 | Filter tasks by priority

As a user, I want to filter the task list to see only high-priority items when I'm focused.

#### Acceptance Criteria

- [ ] Filter dropdown with options: All | High | Medium | Low
- [ ] Filter persists in URL params
- [ ] Empty state message when no tasks match filter
- [ ] Typecheck passes
- [x] Verify in browser using dev-browser skill

## Functional Requirements

- FR-1: Add priority field to tasks table
- FR-2: Display colored priority badge on each task card
- FR-3: Include priority selector in task edit modal

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
*Document Version: 1.1*
*Last Updated: 2026-01-10*
`

func TestParseValidDocument(t *testing.T) {
	f, err := Parse(validDocument)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.Name != "Task Priority System" {
		t.Errorf("Name = %q, want %q", f.Name, "Task Priority System")
	}

	if !strings.Contains(f.Description, "Add priority levels") {
		t.Errorf("Description should contain intro text, got %q", f.Description)
	}

	if !strings.Contains(f.Goals, "Allow assigning priority") {
		t.Errorf("Goals should contain goal text, got %q", f.Goals)
	}

	if len(f.UserStories) != 4 {
		t.Fatalf("expected 4 user stories, got %d", len(f.UserStories))
	}

	us1 := f.UserStories[0]
	if us1.ID != 1 {
		t.Errorf("US[0].ID = %d, want 1", us1.ID)
	}
	if us1.Priority != 1 {
		t.Errorf("US[0].Priority = %d, want 1", us1.Priority)
	}
	if us1.Name != "Add priority field to database" {
		t.Errorf("US[0].Name = %q", us1.Name)
	}
	if len(us1.AcceptanceCriteria) != 3 {
		t.Errorf("US[0] AC count = %d, want 3", len(us1.AcceptanceCriteria))
	}

	us4 := f.UserStories[3]
	if us4.ID != 4 {
		t.Errorf("US[3].ID = %d, want 4", us4.ID)
	}
	if us4.Priority != 2 {
		t.Errorf("US[3].Priority = %d, want 2", us4.Priority)
	}
	if len(us4.AcceptanceCriteria) != 5 {
		t.Errorf("US[3] AC count = %d, want 5", len(us4.AcceptanceCriteria))
	}
	// Last criterion should be completed
	lastAC := us4.AcceptanceCriteria[4]
	if !lastAC.Completed {
		t.Errorf("US[3] last AC should be completed")
	}

	if !strings.Contains(f.FunctionalRequirements, "FR-1") {
		t.Errorf("FunctionalRequirements missing content")
	}

	if !strings.Contains(f.NonGoals, "notifications") {
		t.Errorf("NonGoals missing content")
	}

	if !strings.Contains(f.TechnicalConsiderations, "badge component") {
		t.Errorf("TechnicalConsiderations missing content")
	}

	if !strings.Contains(f.SuccessMetrics, "2 clicks") {
		t.Errorf("SuccessMetrics missing content")
	}

	if f.Version != "1.1" {
		t.Errorf("Version = %q, want %q", f.Version, "1.1")
	}

	if f.LastUpdated != "2026-01-10" {
		t.Errorf("LastUpdated = %q, want %q", f.LastUpdated, "2026-01-10")
	}
}

func TestParseValidDocumentValidates(t *testing.T) {
	f, err := Parse(validDocument)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	errs := Validate(f)
	if len(errs) != 0 {
		t.Errorf("expected no validation errors, got %v", errs)
	}
}

func TestValidateMissingTitle(t *testing.T) {
	f, _ := Parse("## Goals\n\nSome goals\n\n## User Stories\n\n### US-1 | P1 | Story\n\n#### Acceptance Criteria\n\n- [ ] Done\n\n## Functional Requirements\n\nSome requirements\n")
	errs := Validate(f)
	if !hasError(errs, "Name") {
		t.Errorf("expected Name error, got %v", errs)
	}
}

func TestValidateMissingGoals(t *testing.T) {
	f, _ := Parse("# Title\n\n## User Stories\n\n### US-1 | P1 | Story\n\n#### Acceptance Criteria\n\n- [ ] Done\n\n## Functional Requirements\n\nSome requirements\n")
	errs := Validate(f)
	if !hasError(errs, "Goals") {
		t.Errorf("expected Goals error, got %v", errs)
	}
}

func TestValidateNoUserStories(t *testing.T) {
	f, _ := Parse("# Title\n\n## Goals\n\nGoals here\n\n## User Stories\n\n## Functional Requirements\n\nSome requirements\n")
	errs := Validate(f)
	if !hasError(errs, "UserStories") {
		t.Errorf("expected UserStories error, got %v", errs)
	}
}

func TestValidateMissingFunctionalRequirements(t *testing.T) {
	f, _ := Parse("# Title\n\n## Goals\n\nGoals here\n\n## User Stories\n\n### US-1 | P1 | Story\n\n#### Acceptance Criteria\n\n- [ ] Done\n")
	errs := Validate(f)
	if !hasError(errs, "FunctionalRequirements") {
		t.Errorf("expected FunctionalRequirements error, got %v", errs)
	}
}

func TestValidateNonSequentialIDsAllowed(t *testing.T) {
	// Non-sequential IDs should be allowed as long as they're unique
	f, _ := Parse("# Title\n\n## Goals\n\nGoals\n\n## User Stories\n\n### US-3 | P1 | Story\n\n#### Acceptance Criteria\n\n- [ ] Done\n\n## Functional Requirements\n\nReqs\n")
	errs := Validate(f)
	if hasError(errs, "UserStories[0].ID") {
		t.Errorf("non-sequential IDs should be allowed, got %v", errs)
	}
}

func TestValidateDuplicateIDs(t *testing.T) {
	f := Feature{
		Name:                   "Test",
		Goals:                  "Goals",
		FunctionalRequirements: "Reqs",
		UserStories: []UserStory{
			{ID: 1, Priority: 1, Name: "Story 1", AcceptanceCriteria: []AcceptanceCriterion{{Text: "Done"}}},
			{ID: 1, Priority: 2, Name: "Story 2", AcceptanceCriteria: []AcceptanceCriterion{{Text: "Done"}}},
		},
	}
	errs := Validate(f)
	if !hasError(errs, "UserStories[1].ID") {
		t.Errorf("expected duplicate ID error, got %v", errs)
	}
}

func TestValidateMaxUserStories(t *testing.T) {
	f := Feature{
		Name:                   "Test",
		Goals:                  "Goals",
		FunctionalRequirements: "Reqs",
		UserStories:            make([]UserStory, 100),
	}
	for i := range f.UserStories {
		f.UserStories[i] = UserStory{ID: i + 1, Priority: 3, Name: "Story", AcceptanceCriteria: []AcceptanceCriterion{{Text: "Done"}}}
	}
	errs := Validate(f)
	if !hasError(errs, "UserStories") {
		t.Errorf("expected max user stories error, got %v", errs)
	}
}

func TestValidateInvalidPriority(t *testing.T) {
	f := Feature{
		Name:                   "Test",
		Goals:                  "Goals",
		FunctionalRequirements: "Reqs",
		UserStories: []UserStory{
			{ID: 1, Priority: 7, Name: "Story", AcceptanceCriteria: []AcceptanceCriterion{{Text: "Done"}}},
		},
	}
	errs := Validate(f)
	if !hasError(errs, "UserStories[0].Priority") {
		t.Errorf("expected Priority error, got %v", errs)
	}
}

func TestValidateMissingAcceptanceCriteria(t *testing.T) {
	f, _ := Parse("# Title\n\n## Goals\n\nGoals\n\n## User Stories\n\n### US-1 | P1 | Story\n\nSome description but no AC section.\n\n## Functional Requirements\n\nReqs\n")
	errs := Validate(f)
	if !hasError(errs, "UserStories[0].AcceptanceCriteria") {
		t.Errorf("expected AcceptanceCriteria error, got %v", errs)
	}
}

func TestParseDefaultPriority(t *testing.T) {
	f, err := Parse("# T\n\n## Goals\n\nG\n\n## User Stories\n\n### US-1 | Story name\n\n#### Acceptance Criteria\n\n- [ ] Done\n\n## Functional Requirements\n\nR\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.UserStories[0].Priority != 3 {
		t.Errorf("default priority should be 3, got %d", f.UserStories[0].Priority)
	}
	if f.UserStories[0].Name != "Story name" {
		t.Errorf("Name = %q, want %q", f.UserStories[0].Name, "Story name")
	}
}

func TestParseNotes(t *testing.T) {
	f, err := Parse("# T\n\n## Goals\n\nG\n\n## User Stories\n\n### US-1 | P1 | S\n\n#### Acceptance Criteria\n\n- [ ] Done\n\n## Functional Requirements\n\nR\n\n## Notes\n\n### 2026-01-15\n\nSome note content.\n\n### 2026-01-20\n\nAnother note.\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(f.Notes))
	}
	if f.Notes[0].Date != "2026-01-15" {
		t.Errorf("Note[0].Date = %q", f.Notes[0].Date)
	}
	if !strings.Contains(f.Notes[0].Content, "Some note content") {
		t.Errorf("Note[0].Content = %q", f.Notes[0].Content)
	}
}

func TestParseUserStoryTechnicalConsiderations(t *testing.T) {
	f, err := Parse("# T\n\n## Goals\n\nG\n\n## User Stories\n\n### US-1 | P1 | S\n\nDesc.\n\n#### Acceptance Criteria\n\n- [ ] Done\n\n#### Technical Considerations\n\nUse the existing API.\n\n## Functional Requirements\n\nR\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(f.UserStories[0].TechnicalConsiderations, "existing API") {
		t.Errorf("TechnicalConsiderations = %q", f.UserStories[0].TechnicalConsiderations)
	}
}

func hasError(errs []ValidationError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}
