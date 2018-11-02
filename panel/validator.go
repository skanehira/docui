package panel

type Validator struct {
	Message  string
	Validate func(string) bool
}

func NewValidator(msg string, v func(string) bool) *Validator {
	return &Validator{msg, v}
}

var Require = NewValidator("require input", func(text string) bool {
	if text == "" {
		return false
	}

	return true
})
