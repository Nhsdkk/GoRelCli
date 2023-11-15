package schema_parser

import (
	"GoRelCli/utils/env_loader"
	"GoRelCli/utils/error_types/schema_parser_error"
	"GoRelCli/utils/schema_model"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func loadFileFromFS(path string) ([]byte, error) {
	absolutePath, err := filepath.Abs(path)

	if err != nil {
		return nil, schema_parser_error.SchemaParserError{
			Type: schema_parser_error.PathParsingError,
			Text: err.Error(),
		}
	}

	file, err := os.ReadFile(absolutePath)

	if err != nil {
		return nil, schema_parser_error.SchemaParserError{
			Type: schema_parser_error.FileReadingError,
			Text: fmt.Sprintf("Error while reading schema file: %s", err),
		}
	}

	return file, nil
}

func parseYmlFileToSchema(file []byte) (schema_model.GoRelSchema, error) {
	var goRelSchema schema_model.GoRelSchema
	if err := yaml.Unmarshal(file, &goRelSchema); err != nil {
		return schema_model.GoRelSchema{}, schema_parser_error.SchemaParserError{
			Type: schema_parser_error.ParsingError,
			Text: err.Error(),
		}
	}

	return goRelSchema, nil
}

func parseConnectionUrl(schema *schema_model.GoRelSchema) error {
	isEnvFunc, err := regexp.MatchString("^env\\(\\\"\\S*\\\"\\)$", schema.Connection.Url)
	if !isEnvFunc || err != nil {
		if !strings.Contains(schema.Connection.Url, "connect_timeout") {
			schema.Connection.Url += "&connect_timeout=5"
		}
	}

	if err := env_loader.LoadEnvFile(); err != nil {
		return err
	}

	envVariableName := schema.Connection.Url[5 : len(schema.Connection.Url)-2]
	urlEnv, exists := os.LookupEnv(envVariableName)

	if !exists {
		return schema_parser_error.SchemaParserError{
			Type: schema_parser_error.EmptyEnvVariableError,
			Text: fmt.Sprintf("can't find env variable with name %s", envVariableName),
		}
	}

	if !strings.Contains(urlEnv, "connect_timeout") {
		schema.Connection.Url = urlEnv + "&connect_timeout=5"
	} else {
		schema.Connection.Url = urlEnv
	}
	return nil
}

func LoadYmlSchema(path string, value *schema_model.GoRelSchema) error {
	ymlFile, err := loadFileFromFS(path)
	if err != nil {
		return err
	}
	goRelSchema, err := parseYmlFileToSchema(ymlFile)
	*value = goRelSchema
	if err != nil {
		return err
	}
	if err := parseConnectionUrl(value); err != nil {
		return err
	}
	return nil
}

func IndexSchema(schema schema_model.GoRelSchema) (enumNames []string, modelNames []string) {
	modelNames = make([]string, len(schema.Models))
	enumNames = make([]string, len(schema.Enums))
	for index, model := range schema.Models {
		modelNames[index] = model.Name
	}
	for index, enum := range schema.Enums {
		enumNames[index] = enum.Name
	}
	return
}
