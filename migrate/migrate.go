package migrate

import (
	"GoRelCli/logger"
	"GoRelCli/migrate/database_contoller"
	"GoRelCli/schema_model"
	"GoRelCli/schema_parser"
	"GoRelCli/vaidator"
	"bufio"
	"database/sql"
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
		enumNamesInn, modelNamesInn, err := vaidator.ValidateSchema(&goRelSchema)
		if err != nil {
			return err
		}
		enumNames = enumNamesInn
		modelNames = modelNamesInn
		return nil
	}); err != nil {
		return err
	}

	var db *sql.DB

	if err := logger.LogStep("connect to db", func() error {
		dbClientPtr, err := database_contoller.GetController(goRelSchema.Connection.Provider, goRelSchema.Connection.Url)
		db = dbClientPtr

		if err != nil {
			return errors.New(fmt.Sprintf("Error while connecting to the db:\n%s", err))
		}

		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("drop tables", func() error {
		if err := database_contoller.DropTables(db); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("drop enums", func() error {
		if err := database_contoller.DropEnums(db); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("create enums", func() error {
		if err := database_contoller.CreateEnums(db, goRelSchema.Enums); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := logger.LogStep("create tables", func() error {
		if err := database_contoller.CreateTables(db, goRelSchema.Models, enumNames, modelNames); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
