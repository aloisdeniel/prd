// / Package prd defines the data structures for representing a parsed PRD document.
package prd

// Feature represents a parsed PRD markdown document.
type Feature struct {
	Name                    string
	Description             string
	Goals                   string
	UserStories             []UserStory
	FunctionalRequirements  string
	NonGoals                string
	TechnicalConsiderations string
	Analytics               string
	Notes                   []Note
	Risks                   string
	SuccessMetrics          string
	OpenQuestions           string
	Version                 string
	LastUpdated             string
}

// UserStory represents a single user story within a feature.
type UserStory struct {
	ID                      int
	Priority                int
	Name                    string
	Description             string
	AcceptanceCriteria      []AcceptanceCriterion
	TechnicalConsiderations string
}

// AcceptanceCriterion represents a single checklist item.
type AcceptanceCriterion struct {
	Text      string
	Completed bool
}

// FeatureEntry represents a feature in the listing with summary info.
type FeatureEntry struct {
	ID        string
	Name      string
	Path      string
	Completed int
	Total     int
}

// Note represents a dated note entry.
type Note struct {
	Date    string
	Content string
}
