package panel

// Validator validator.
type Validator struct {
	Message  string
	Validate func(string) bool
}

// NewValidator validate input.
func NewValidator(msg string, v func(string) bool) *Validator {
	return &Validator{msg, v}
}

// Require require validation.
var Require = NewValidator("no specified ", func(text string) bool {
	if text == "" {
		return false
	}

	return true
})
