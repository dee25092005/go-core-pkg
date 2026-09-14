package apperrors

type FieldViolation struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
	Error  string `json:"error"`
}

type ValidationErrorBuilder struct {
	violations []FieldViolation
}

func Validation() *ValidationErrorBuilder {
	return &ValidationErrorBuilder{
		violations: make([]FieldViolation, 0),
	}
}

func (v *ValidationErrorBuilder) Add(
	field string,
	reason string,
	err string,
) *ValidationErrorBuilder {
	v.violations = append(v.violations, FieldViolation{
		Field:  field,
		Reason: reason,
		Error:  err,
	})

	return v
}

func ValidationErrors(violations []FieldViolation) *AppError {
	err := BadRequest("Validation failed")

	err.Detail = violations

	return err
}

func (v *ValidationErrorBuilder) Error() *AppError {
	if len(v.violations) == 0 {
		return nil
	}

	return &AppError{
		Code:    400,
		Message: "Validation failed",
		Status:  "INVALID_ARGUMENT",
		Detail:  v.violations,
	}
}
