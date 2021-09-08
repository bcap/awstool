package common

import (
	"fmt"
	"strings"
)

type Errors struct {
	Errors []error
}

func NewErrors(errors []error) Errors {
	return Errors{Errors: errors}
}

func (e Errors) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	errorStrings := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		errorStrings[i] = "\"" + err.Error() + "\""
	}
	return fmt.Sprintf("multiple errors: [%s]", strings.Join(errorStrings, ", "))
}
