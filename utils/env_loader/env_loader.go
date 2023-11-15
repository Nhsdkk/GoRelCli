package env_loader

import (
	"GoRelCli/utils/error_types/env_loader_error"
	"bufio"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

func LoadEnvFile() error {
	fmt.Println("Specify path to .env file (relative path only):")
	reader := bufio.NewReader(os.Stdin)
	relativePath, err := reader.ReadString('\n')

	if err != nil {
		return env_loader_error.EnvLoaderError{
			Type: env_loader_error.ReadingFromStdioError,
			Text: "Error while reading from stdio",
		}
	}

	relativePath = relativePath[0 : len(relativePath)-2]
	absolutePath, err := filepath.Abs(relativePath)

	if err != nil {
		return env_loader_error.EnvLoaderError{
			Type: env_loader_error.ResolvingPathError,
			Text: fmt.Sprintf("Can't resolve path \"%s\"", relativePath),
		}
	}

	if err := godotenv.Load(absolutePath); err != nil {
		return env_loader_error.EnvLoaderError{
			Type: env_loader_error.ReadingEnvFileError,
			Text: fmt.Sprintf("Can't read file from path \"%s\"", absolutePath),
		}
	}

	return nil
}
