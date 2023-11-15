package generate

import (
	"GoRelCli/models/schema_model"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type GoRelGeneratedFileImpl struct {
	absolutePath string
	content      string
	fileType     FileType
}

type FileType string

type ObjectUnionType struct {
	model    schema_model.Model
	enum     schema_model.Enum
	fileType FileType
}

const (
	MODEL FileType = "Models"
	ENUM           = "Enums"
)

func (g *GoRelGeneratedFileImpl) Create(object ObjectUnionType, enumNames []string, modelNames []string, projectName string, projectPath string) error {
	g.fileType = object.fileType
	if err := g.generateFileContent(object, enumNames, modelNames, projectName); err != nil {
		return err
	}

	relativePath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	if g.fileType == MODEL {
		g.absolutePath = fmt.Sprintf("%s\\gorel\\models\\%s.go", relativePath, object.model.Name)
	} else {
		g.absolutePath = fmt.Sprintf("%s\\gorel\\enums\\%s.go", relativePath, object.enum.Name)
	}

	return nil
}

func (g *GoRelGeneratedFileImpl) WriteFS() (err error) {
	file, err := os.OpenFile(g.absolutePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		if errInn := file.Close(); errInn != nil {
			err = errInn
		}
	}(file)

	byteRepr := []byte(g.content)
	n, err := file.Write(byteRepr)
	if err != nil {
		return err
	}

	if n != len(byteRepr) {
		return errors.New(fmt.Sprintf("wrote %d bytes, but content length is %d", n, len(byteRepr)))
	}
	return err
}

func (g *GoRelGeneratedFileImpl) Log() {
	fmt.Println(fmt.Sprintf("Contents of %s file with path %s:\n%s", g.fileType, g.absolutePath, g.content))
}

func (g *GoRelGeneratedFileImpl) createFile(filename string) (*os.File, error) {
	return os.Create(filename)
}

func (g *GoRelGeneratedFileImpl) writeFile(file *os.File, text string) error {
	byteRepr := []byte(text)
	bytesWritten, err := file.Write(byteRepr)
	if err != nil || bytesWritten != len(byteRepr) {
		return errors.New(fmt.Sprintf("Error while writing to file %s", file.Name()))
	}
	return nil
}

func (g *GoRelGeneratedFileImpl) generateFileContent(object ObjectUnionType, enumNames []string, modelNames []string, projectName string) error {
	var structString string
	var err error
	var referenceModels, referenceEnums []string

	if object.fileType == MODEL {
		structString, referenceModels, referenceEnums, err = g.generateStructModel(object.model, enumNames, modelNames)
	} else {
		structString = g.generateEnum(object.enum)
	}

	if err != nil {
		return err
	}
	importString := g.generateImports(referenceEnums, referenceModels, projectName)
	g.content = fmt.Sprintf("%s\n%s", importString, structString)
	return nil
}

func (g *GoRelGeneratedFileImpl) generateImports(referenceEnums []string, referenceModels []string, projectName string) string {
	importString := ""

	if g.fileType == MODEL {
		importString = "package models\n\n"
	} else {
		importString = "package enums\n\n"
	}

	if len(referenceEnums) != 0 && g.fileType != ENUM {
		importString += fmt.Sprintf("import \"%s/gorel/enums\"\n", projectName)
	}

	if len(referenceModels) != 0 && g.fileType != MODEL {
		importString += fmt.Sprintf("import \"%s/gorel/models\"\n", projectName)
	}

	return importString
}

func (g *GoRelGeneratedFileImpl) generateStructModel(model schema_model.Model, enumNames []string, modelNames []string) (structString string, referenceModels []string, referenceEnums []string, err error) {
	structString = fmt.Sprintf("type %s struct{\n", model.Name)
	for _, property := range model.Properties {
		if property.RelationField != "" {
			continue
		}
		goLangType, isValidGoLangType := property.GetGoLangType()
		if !isValidGoLangType {
			if slices.Contains(enumNames, goLangType) {
				referenceEnums = append(referenceEnums, property.Type)
				structString += fmt.Sprintf("\t%s enums.%s\n", property.Name, goLangType)
				continue
			}

			clearedModelName := ""
			if strings.Contains(goLangType, "[]") {
				clearedModelName = goLangType[:len(goLangType)-2]
			} else {
				clearedModelName = goLangType
			}

			if slices.Contains(modelNames, clearedModelName) {
				referenceModels = append(referenceModels, property.Type)
				structString += fmt.Sprintf("\t%s []%s\n", property.Name, clearedModelName)
				continue
			}

			return "", nil, nil, errors.New(fmt.Sprintf("Property with name %s has wrong type %s", property.Name, property.Type))
		}
		structString += fmt.Sprintf("\t%s %s\n", property.Name, goLangType)
	}
	structString += "}"
	return structString, referenceModels, referenceEnums, nil
}

func (g *GoRelGeneratedFileImpl) generateEnum(enum schema_model.Enum) string {
	enumString := fmt.Sprintf("type %s string\n\nconst (\n", enum.Name)
	for i, value := range enum.Values {
		if i == 0 {
			enumString += fmt.Sprintf("\t%s %s = \"%s\"\n", value, enum.Name, strings.ToUpper(value))
			continue
		}
		enumString += fmt.Sprintf("\t%s = \"%s\"\n", value, strings.ToUpper(value))
	}
	enumString += ")"
	return enumString
}
