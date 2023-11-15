package generate

import (
	"GoRelCli/logger"
	"GoRelCli/schema_model"
	"GoRelCli/schema_parser"
	"GoRelCli/validator"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getPathArguments(args ...string) (schemaPath string, projectPath string, err error) {
	if args[0] == "" || args[1] == "" {
		return "", "", errors.New("not enough parameters")
	}

	projectPath, err = filepath.Abs(args[1])
	if err != nil {
		return "", "", err
	}

	return args[0], projectPath, nil
}

func getProjectName(path string) (string, error) {
	if !strings.Contains(path, "\\") {
		return "", errors.New("can't get project name from path")
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	splittedPath := strings.Split(absolutePath, "\\")
	return splittedPath[len(splittedPath)-1], nil
}

func createFileObjects(schema schema_model.GoRelSchema, modelNames []string, enumNames []string, projectName string, projectPath string) ([]GoRelGeneratedFileInterface, error) {
	var fileObjects []GoRelGeneratedFileInterface
	for _, model := range schema.Models {
		object := ObjectUnionType{
			fileType: MODEL,
			model:    model,
		}
		fileObject := GoRelGeneratedFileImpl{}
		if err := fileObject.Create(object, enumNames, modelNames, projectName, projectPath); err != nil {
			return nil, err
		}
		fileObjects = append(fileObjects, &fileObject)
		fileObject.Log()
	}

	for _, enum := range schema.Enums {
		object := ObjectUnionType{
			fileType: ENUM,
			enum:     enum,
		}
		fileObject := GoRelGeneratedFileImpl{}
		if err := fileObject.Create(object, enumNames, modelNames, projectName, projectPath); err != nil {
			return nil, err
		}
		fileObjects = append(fileObjects, &fileObject)
		fileObject.Log()
	}

	return fileObjects, nil
}

func checkFolder(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func checkFolders(modelNames []string, enumNames []string, projectPath string) error {
	filePath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	var folderPath string

	if len(modelNames) != 0 {
		folderPath = fmt.Sprintf("%s\\gorel\\models", filePath)
		if isValid, err := checkFolder(folderPath); err != nil {
			return err
		} else if !isValid {
			if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
				return err
			}
		}
	}
	if len(enumNames) != 0 {
		folderPath = fmt.Sprintf("%s\\gorel\\enums", filePath)
		if isValid, err := checkFolder(folderPath); err != nil {
			return err
		} else if !isValid {
			if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
				return err
			}
		}
	}
	if len(enumNames) == 0 && len(modelNames) == 0 {
		return errors.New("no enums nor models to create")
	}

	return nil
}

//func createFileAsync(filename string,rootPath string, c chan FileCreateResult, syncGroup *sync.WaitGroup) {
//	path := fmt.Sprintf("%s/%s.go", rootPath, filename)
//	go func(path string, c chan FileCreateResult) {
//		defer syncGroup.Done()
//		file, err := createFile(path)
//		defer func(file *os.File) {
//			err := file.Close()
//			if err != nil {
//				c <- FileCreateResult{
//					path: "",
//					err: nil,
//				}
//			}
//			c <- FileCreateResult{
//				filename: filename,
//				path:    path,
//				err:     err,
//			}
//		}(file)
//	}(path, c)
//}
//
//func createFiles(modelNames []string, enumNames []string, rootPath string) (map[string]*os.File, error) {
//	var syncGroup sync.WaitGroup
//	fileMapping := make(map[string]*os.File)
//
//	c := make(chan FileCreateResult)
//	defer close(c)
//
//	for _, modelName := range modelNames {
//		syncGroup.Add(1)
//		go createFileAsync(, c, &syncGroup)
//	}
//
//	for _, enumName := range enumNames {
//		syncGroup.Add(1)
//		go createFileAsync(fmt.Sprintf("%s/%s.go", rootPath, enumName), c, &syncGroup)
//	}
//
//	syncGroup.Wait()
//
//	for res := range c {
//		if res.err != nil {
//			return nil, res.err
//		}
//		fileMapping[] = res.filePtr
//	}
//
//	return fileMapping, nil
//}

func createFiles(fileObjects []GoRelGeneratedFileInterface) error {
	for _, fileObject := range fileObjects {
		if err := fileObject.WriteFS(); err != nil {
			return err
		}
	}
	return nil
}

func Generate(args ...string) error {
	var schemaPath, projectPath string

	if err := logger.LogStep("get project name", func() error {
		schemaPathInn, projectPathInn, err := getPathArguments(args...)
		if err != nil {
			return err
		}
		schemaPath = schemaPathInn
		projectPath = projectPathInn
		return nil
	}); err != nil {
		return err
	}

	var goRelSchema schema_model.GoRelSchema
	var projectName string

	if err := logger.LogStep("get project name", func() error {
		projectNameInn, err := getProjectName(projectPath)
		if err != nil {
			return err
		}
		projectName = projectNameInn
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("load schema", func() error {
		if err := schema_parser.LoadYmlSchema(schemaPath, &goRelSchema); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	var enumNames []string
	var modelNames []string

	if err := logger.LogStep("validate schema", func() error {
		enumNamesInn, modelNamesInn, err := validator.ValidateSchema(&goRelSchema)
		if err != nil {
			return err
		}
		enumNames, modelNames = enumNamesInn, modelNamesInn
		return nil
	}); err != nil {
		return err
	}

	var fileObjects []GoRelGeneratedFileInterface

	if err := logger.LogStep("generate file objects", func() error {
		fileObjectsInn, err := createFileObjects(goRelSchema, modelNames, enumNames, projectName, projectPath)
		if err != nil {
			return err
		}
		fileObjects = fileObjectsInn
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("check folders existance", func() error {
		if err := checkFolders(modelNames, enumNames, projectPath); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("create files", func() error {
		if err := createFiles(fileObjects); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
