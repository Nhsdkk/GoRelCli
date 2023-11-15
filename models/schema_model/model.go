package schema_model

import (
	"GoRelCli/models/error_model/validation_error"
	"fmt"
	"strconv"
)

type Model struct {
	Name       string     `yaml:"name"`
	Properties []Property `yaml:"properties,flow"`
}

type Property struct {
	Name           string `yaml:"name"`
	Type           string `yaml:"type"`
	Default        string `yaml:"default,omitempty"`
	Unique         bool   `yaml:"unique"`
	Id             bool   `yaml:"id"`
	RelationField  string `yaml:"relationField,omitempty"`
	ReferenceField string `yaml:"referenceField,omitempty"`
}

func (p *Property) GetPostgresType() (postgresType string, isValidPostgresType bool) {
	typed := PropertyType(p.Type)
	postgresType = postgresTypes[typed]
	if postgresType == "" {
		return p.Type, false
	}
	return postgresType, true
}

func (p *Property) GetGoLangType() (goType string, isValidGoType bool) {
	typed := PropertyType(p.Type)
	goType = goTypes[typed]
	if goType == "" {
		return p.Type, false
	}
	return goType, true
}

func (p *Property) ValidateDefaultValue() (value any, err *validation_error.ValidationError) {
	typed := PropertyType(p.Type)

	switch typed {
	case Int:
		if p.Default == "autoincrement()" {
			return nil, nil
		}
		val, err := strconv.ParseInt(p.Default, 10, 64)
		if err != nil {
			return nil, &validation_error.ValidationError{
				Position: validation_error.ModelValidationError,
				Text:     fmt.Sprintf("%s property. Can't parse default int value from \"%s\"", p.Name, p.Default),
			}
		}
		return val, nil
	case String:
		if p.Default == "uuid()" {
			return nil, nil
		}
		return p.Default, nil
	case Float:
		val, err := strconv.ParseFloat(p.Default, 64)
		if err != nil {
			return nil, &validation_error.ValidationError{
				Position: validation_error.ModelValidationError,
				Text:     fmt.Sprintf("%s property. Can't parse default float value from \"%s\"", p.Name, p.Default),
			}
		}
		return val, nil
	case Boolean:
		val, err := strconv.ParseBool(p.Default)
		if err != nil {
			return nil, &validation_error.ValidationError{
				Position: validation_error.ModelValidationError,
				Text:     fmt.Sprintf("%s property. Can't parse default boolean value from \"%s\"", p.Name, p.Default),
			}
		}
		return val, nil
	case DateTime:
		if p.Default == "now()" {
			return nil, nil
		}
		return nil, &validation_error.ValidationError{
			Position: validation_error.ModelValidationError,
			Text:     fmt.Sprintf("%s property. Can't use dateTime default value of type %s if it's not now()", p.Name, p.Default),
		}
	default:
		return nil, &validation_error.ValidationError{
			Position: validation_error.ModelValidationError,
			Text:     fmt.Sprintf("%s property. Can't use variable of type %s as default value", p.Name, typed),
		}
	}
}

type PropertyType string

const (
	Int              PropertyType = "int"
	Boolean                       = "boolean"
	Float                         = "float"
	String                        = "string"
	DateTime                      = "dateTime"
	IntArr                        = "int[]"
	BooleanArr                    = "boolean[]"
	FloatArr                      = "float[]"
	StringArr                     = "string[]"
	DateTimeArr                   = "dateTime[]"
	IntNullable                   = "int?"
	BooleanNullable               = "boolean?"
	FloatNullable                 = "float?"
	StringNullable                = "string?"
	DateTimeNullable              = "dateTime?"
)

var (
	postgresTypes = map[PropertyType]string{
		Int:              "int NOT NULL",
		Boolean:          "boolean NOT NULL",
		Float:            "double precision NOT NULL",
		String:           "text NOT NULL",
		DateTime:         "timestamptz NOT NULL",
		IntArr:           "int[]",
		BooleanArr:       "boolean[]",
		FloatArr:         "double precision[]",
		StringArr:        "text[]",
		DateTimeArr:      "timestamptz[]",
		IntNullable:      "int",
		BooleanNullable:  "boolean",
		FloatNullable:    "double precision",
		StringNullable:   "text",
		DateTimeNullable: "timestamptz",
	}
	goTypes = map[PropertyType]string{
		Int:              "int64",
		Boolean:          "bool",
		Float:            "float64",
		String:           "string",
		DateTime:         "time",
		IntArr:           "[]int",
		BooleanArr:       "[]boolean",
		FloatArr:         "[]float",
		StringArr:        "[]string",
		DateTimeArr:      "[]time",
		IntNullable:      "int64",
		BooleanNullable:  "bool",
		FloatNullable:    "float64",
		StringNullable:   "string",
		DateTimeNullable: "time",
	}
)
