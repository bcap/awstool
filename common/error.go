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
	errorStrings := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Sprintf("multiple errors: [%s]", strings.Join(errorStrings, ","))
}
