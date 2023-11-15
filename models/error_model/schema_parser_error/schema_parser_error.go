package schema_parser_error

import "fmt"

type ErrorType string

const (
	ParsingError          ErrorType = "parsing error"
	EmptyEnvVariableError           = "env variable does not exist error"
	FileReadingError                = "file not exists error"
	PathParsingError                = "path parsing error"
)

type SchemaParserError struct {
	Type ErrorType
	Text string
}

func (e SchemaParserError) Error() string {
	return fmt.Sprintf("Error while parsing schema:\n%s - %s", e.Type, e.Text)
}
