package migrate

import (
	"GoRelCli/logger"
	"GoRelCli/migrate/database_contoller"
	"GoRelCli/schema_model"
	"GoRelCli/schema_parser"
	"GoRelCli/validator"
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func checkFlags(args ...string) (valid bool) {
	if args[0] == "" {
		return false
	}
	return true
}

func requestPermissionToOverrideSchema() error {
	fmt.Println("This action will delete all existing enums and tables. Are you sure you want to proceed? (Y-yes/N-no):")
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	switch strings.ToLower(str[0:1]) {
	case "y":
		return nil
	case "n":
		return errors.New("user refused to give permission to override tables and enums")
	default:
		return errors.New("unknown option")
	}
}

func Migrate(args ...string) error {
	if err := requestPermissionToOverrideSchema(); err != nil {
		return errors.New(fmt.Sprintf("error while requesting permission to override:\n\t%s", err))
	}
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

	var modelNames []string
	var enumNames []string

	if err := logger.LogStep("validate schema", func() error {
		enumNamesInn, modelNamesInn, err := validator.ValidateSchema(&goRelSchema)
		if err != nil {
			return err
		}
		enumNames = enumNamesInn
		modelNames = modelNamesInn
		return nil
	}); err != nil {
		return err
	}

	var databaseController database_contoller.DatabaseControllerInterface

	if err := logger.LogStep("connect to db", func() error {
		databaseControllerInner, err := database_contoller.NewDatabaseController(goRelSchema.Connection)
		if err != nil {
			return err
		}
		databaseController = databaseControllerInner
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("run migrations", func() error {
		err := databaseController.RunMigrations(&goRelSchema, enumNames, modelNames)
		return err
	}); err != nil {
		return err
	}

	return nil
}
