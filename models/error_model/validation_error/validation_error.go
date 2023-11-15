package validation_error

import "fmt"

type ValidationErrorPosition string

const (
	ModelValidationError    ValidationErrorPosition = "model validation error"
	EnumValidationError                             = "enum validation error"
	RelationValidationError                         = "relation validation error"
)

type ValidationError struct {
	Position ValidationErrorPosition
	Text     string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("Error while validating file:\n%s - %s", e.Position, e.Text)
}
