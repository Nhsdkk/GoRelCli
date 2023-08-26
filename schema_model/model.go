package schema_model

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
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
	typed := propertyType(p.Type)
	postgresType = postgresTypes[typed]
	if postgresType == "" {
		return p.Type, false
	}
	return postgresType, true
}

func (p *Property) ValidateDefaultValue() (any, error) {
	typed := propertyType(p.Type)
	switch typed {
	case Int:
		val, err := strconv.ParseInt(p.Default, 10, 64)
		return val, err
	case String:
		return p.Default, nil
	case Float:
		val, err := strconv.ParseFloat(p.Default, 64)
		return val, err
	case Boolean:
		val, err := strconv.ParseBool(p.Default)
		return val, err
	case DateTime:
		return nil, errors.New("can't use dateTime default value if it's not now()")
	default:
		return nil, errors.New(fmt.Sprintf("can't use variable of type %s as default value", typed))
	}
}

type propertyType string

const (
	Int              propertyType = "int"
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
	postgresTypes = map[propertyType]string{
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
	goTypes = map[propertyType]reflect.Type{
		Int:      reflect.TypeOf(1),
		Boolean:  reflect.TypeOf(true),
		Float:    reflect.TypeOf(1.2),
		String:   reflect.TypeOf(" "),
		DateTime: reflect.TypeOf(time.Now()),
		//IntArr:           reflect.TypeOf([1,2]),
		//BooleanArr:       reflect.TypeOf(),
		//FloatArr:         reflect.TypeOf(),
		//StringArr:        reflect.TypeOf(),
		//DateTimeArr:      reflect.TypeOf(),
	}
)
