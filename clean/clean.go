package clean

import (
	"GoRelCli/models/schema_model"
	"GoRelCli/utils/logger"
	"GoRelCli/utils/schema_parser"
	"GoRelCli/utils/validator"
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func checkFlags(args ...string) (valid bool) {
	if args[0] == "" {
		return false
	}
	return true
}

func writeYmlFS(schema schema_model.GoRelSchema, path string) (err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		if errInn := file.Close(); errInn != nil {
			err = errInn
		}
	}(file)

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(schema); err != nil {
		return err
	}

	return err
}

func Clean(args ...string) error {
	if !checkFlags(args...) {
		return errors.New("path flag should be provided")
	}
	path := args[0]

	var goRelSchema schema_model.GoRelSchema

	if err := logger.LogStep("load schema", func() error {
		if err := schema_parser.LoadYmlSchema(path, &goRelSchema); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("validate schema", func() error {
		_, _, err := validator.ValidateSchema(&goRelSchema)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		if err := logger.LogStep("cleanup names inside schema", func() error {
			validator.CleanupNames(&goRelSchema)
			return nil
		}); err != nil {
			return nil
		}
	}

	if err := logger.LogStep("write schema to fs", func() error {
		if err := writeYmlFS(goRelSchema, path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
