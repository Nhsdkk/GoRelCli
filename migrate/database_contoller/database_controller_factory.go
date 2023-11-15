package database_contoller

import (
	"GoRelCli/utils/error_types/database_error"
	"GoRelCli/utils/schema_model"
	"database/sql"
	"fmt"
)

func getPostgreSQLDatabaseController(url string) (*PostgresController, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, database_error.DatabaseError{
			ErrorType: database_error.ConnectionError,
			Text:      fmt.Sprintf("Can't connect to database with url: %s", url),
		}
	}
	controller := &PostgresController{db: db}
	if err := controller.checkConnection(); err != nil {
		return nil, err
	}
	return &PostgresController{db: db}, nil
}

func NewDatabaseController(connectionInfo schema_model.Connection) (DatabaseControllerInterface, error) {
	switch connectionInfo.Provider {
	case schema_model.PostgreSQL:
		controller, err := getPostgreSQLDatabaseController(connectionInfo.Url)
		if err != nil {
			return nil, err
		}

		return controller, nil
	case schema_model.MySQL:
		return nil, database_error.DatabaseError{
			ErrorType: database_error.UnsupportedProviderError,
			Text:      fmt.Sprintf("%s is not supported.", connectionInfo.Provider),
		}
	default:
		return nil, database_error.DatabaseError{
			ErrorType: database_error.UnsupportedProviderError,
			Text:      fmt.Sprintf("%s is not supported.", connectionInfo.Provider),
		}
	}

}
