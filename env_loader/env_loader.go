package env_loader

import (
	"bufio"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

type EnvLoaderErrorPosition string

const (
	ReadingFromStdio EnvLoaderErrorPosition = "reading from stdio"
	ResolvingPath                           = "resolving path"
	ReadingEnvFile                          = "reading env file from fs"
)

type EnvLoaderError struct {
	position EnvLoaderErrorPosition
	text     string
}

func (e EnvLoaderError) Error() string {
	return fmt.Sprintf("Error while %s: %s", e.position, e.text)
}

func LoadEnvFile() *EnvLoaderError {
	fmt.Println("Specify path to .env file (relative path only):")
	reader := bufio.NewReader(os.Stdin)
	relativePath, err := reader.ReadString('\n')

	if err != nil {
		return &EnvLoaderError{
			position: ReadingFromStdio,
			text:     "Error while reading from stdio",
		}
	}

	relativePath = relativePath[0 : len(relativePath)-2]
	absolutePath, err := filepath.Abs(relativePath)

	if err != nil {
		return &EnvLoaderError{
			position: ResolvingPath,
			text:     fmt.Sprintf("Can't resolve path \"%s\"", relativePath),
		}
	}

	if err := godotenv.Load(absolutePath); err != nil {
		return &EnvLoaderError{
			position: ReadingEnvFile,
			text:     fmt.Sprintf("Can't read file from path \"%s\"", absolutePath),
		}
	}

	return nil
}
