package generate

import "sync"

type GoRelGeneratedFileInterface interface {
	Create(object ObjectUnionType, enumNames []string, modelNames []string, projectName string, projectPath string) error
	WriteFS() error
	WriteFSAsync(c chan error, syncGroup *sync.WaitGroup)
	Log()
}
