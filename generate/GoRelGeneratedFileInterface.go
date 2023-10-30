package generate

type GoRelGeneratedFileInterface interface {
	Create(object ObjectUnionType, enumNames []string, modelNames []string, projectName string, projectPath string) error
	WriteFS() error
	Log()
}
