package database_contoller

import (
	"GoRelCli/env_loader"
	"GoRelCli/schema_model"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
)

func getPostgreSQLDatabaseController(url string) (*PostgresController, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &PostgresController{db: db}, nil
}

func NewDatabaseController(connectionInfo schema_model.Connection) (DatabaseControllerInterface, error) {
	switch connectionInfo.Provider {
	case schema_model.PostgreSQL:
		isEnvFunc, err := regexp.MatchString("^env\\(\\\"\\S*\\\"\\)$", connectionInfo.Url)

		if err != nil || !isEnvFunc {
			controller, err := getPostgreSQLDatabaseController(connectionInfo.Url)
			if err != nil {
				return nil, err
			}

			return controller, nil
		} else {
			if err := env_loader.LoadEnvFile(); err != nil {
				return nil, err
			}

			envVariableName := connectionInfo.Url[5 : len(connectionInfo.Url)-2]
			urlEnv, exists := os.LookupEnv(envVariableName)

			if !exists {
				return nil, errors.New(fmt.Sprintf("can't find env variable with name %s", envVariableName))
			}

			controller, err := getPostgreSQLDatabaseController(urlEnv)
			if err != nil {
				return nil, err
			}

			return controller, nil
		}
	default:
		return nil, DatabaseError{
			errorType: UnsupportedProviderError,
			text:      fmt.Sprintf("%s is not supported.", connectionInfo.Provider),
		}
	}

}
