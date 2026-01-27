package prd

import "fmt"

// ValidationError describes a single validation issue.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks the Feature for required fields and structural correctness.
func Validate(f Feature) []ValidationError {
	var errs []ValidationError

	if f.Name == "" {
		errs = append(errs, ValidationError{Field: "Name", Message: "H1 title is required"})
	}

	if f.Goals == "" {
		errs = append(errs, ValidationError{Field: "Goals", Message: "Goals section is required and must not be empty"})
	}

	if len(f.UserStories) == 0 {
		errs = append(errs, ValidationError{Field: "UserStories", Message: "at least one user story is required"})
	}

	for i, us := range f.UserStories {
		prefix := fmt.Sprintf("UserStories[%d]", i)

		expectedID := i + 1
		if us.ID != expectedID {
			errs = append(errs, ValidationError{
				Field:   prefix + ".ID",
				Message: fmt.Sprintf("expected US-%d but got US-%d", expectedID, us.ID),
			})
		}

		if us.Priority < 1 || us.Priority > 5 {
			errs = append(errs, ValidationError{
				Field:   prefix + ".Priority",
				Message: fmt.Sprintf("priority must be between 1 and 5, got %d", us.Priority),
			})
		}

		if len(us.AcceptanceCriteria) == 0 {
			errs = append(errs, ValidationError{
				Field:   prefix + ".AcceptanceCriteria",
				Message: "at least one acceptance criterion is required",
			})
		}
	}

	if f.FunctionalRequirements == "" {
		errs = append(errs, ValidationError{Field: "FunctionalRequirements", Message: "Functional Requirements section is required and must not be empty"})
	}

	return errs
}
