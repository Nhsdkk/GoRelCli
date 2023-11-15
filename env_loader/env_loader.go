package env_loader

import (
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
		return err
	}
	relativePath = relativePath[0 : len(relativePath)-2]
	absolutePath, err := filepath.Abs(relativePath)

	if err != nil {
		return err
	}
	if err := godotenv.Load(absolutePath); err != nil {
		return err
	}
	return nil
}
