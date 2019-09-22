package utils

import (
	"fmt"
	"strings"
)

type EnumSlice struct {
	ValidValues []string
	CValues     *[]string
	Name        string
}

func NewEnumSlice(values *[]string, validvalues []string, name string) *EnumSlice {
	return &EnumSlice{
		ValidValues: validvalues,
		CValues:     values,
		Name:        name,
	}
}

func (e *EnumSlice) Values() string {
	return strings.Join(e.ValidValues, ", ")
}

func (e *EnumSlice) Set(value string) error {
	if e.CValues == nil {
		*e.CValues = make([]string, 0, 5)
	}

	for _, v := range e.ValidValues {
		if strings.EqualFold(v, value) {
			*e.CValues = append(*e.CValues, v)
			return nil
		}
	}
	return fmt.Errorf("invalid value '%s', must be one of %s", value, e.Values())
}

func (e *EnumSlice) String() string {
	return strings.Join(*e.CValues, ",")
}

func (e *EnumSlice) Type() string {
	return e.Name
}

type Enum struct {
	ValidValues []string
	Value       *string
	Name        string
}

func NewEnum(value *string, values []string, name string) *Enum {
	return &Enum{
		ValidValues: values,
		Value:       value,
		Name:        name,
	}
}

func (e *Enum) Values() string {
	return strings.Join(e.ValidValues, ", ")
}

func (e *Enum) Set(value string) error {
	for _, v := range e.ValidValues {
		if strings.EqualFold(v, value) {
			*e.Value = v
			return nil
		}
	}
	return fmt.Errorf("invalid value '%s', must be one of %s", value, e.Values())
}

func (e *Enum) String() string {
	return *e.Value
}

func (e *Enum) Type() string {
	return e.Name
}
