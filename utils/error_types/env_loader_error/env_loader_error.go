package env_loader_error

import "fmt"

type EnvLoaderErrorType string

const (
	ReadingFromStdioError EnvLoaderErrorType = "reading from stdio error"
	ResolvingPathError                       = "resolving path error"
	ReadingEnvFileError                      = "reading env file from fs error"
)

type EnvLoaderError struct {
	Type EnvLoaderErrorType
	Text string
}

func (e EnvLoaderError) Error() string {
	return fmt.Sprintf("Error while using env file:\n%s - %s", e.Type, e.Text)
}
