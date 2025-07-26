package flags

import (
	"fmt"
	"slices"
)

// ChoiceValue implements the pflag.Value interface.
type ChoiceValue struct {
	Value string

	validate func(string) error
}

func (f *ChoiceValue) Set(s string) error {
	err := f.validate(s)
	if err != nil {
		return err
	}

	f.Value = s
	return nil
}

func (f *ChoiceValue) Type() string { return "string" }

func (f *ChoiceValue) String() string {
	if f == nil {
		return ""
	}

	return f.Value
}

// StringChoice returns a [choiceValue] that validates the value against a set
// of choices.
func StringChoice(choices []string) *ChoiceValue {
	return &ChoiceValue{
		validate: func(s string) error {
			if slices.Contains(choices, s) {
				return nil
			}

			return fmt.Errorf("must be one of %v", choices)
		},
	}
}
