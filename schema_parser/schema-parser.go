package schema_parser

import (
	"GoRelCli/schema_model"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func loadFileFromFS(path string) ([]byte, error) {
	absolutePath, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(absolutePath)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func parseYmlFileToSchema(file []byte) (schema_model.GoRelSchema, error) {
	var goRelSchema schema_model.GoRelSchema
	if err := yaml.Unmarshal(file, &goRelSchema); err != nil {
		return schema_model.GoRelSchema{}, err
	}

	return goRelSchema, nil
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
