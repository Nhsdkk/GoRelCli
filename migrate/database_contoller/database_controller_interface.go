package database_contoller

import (
	schema_model2 "GoRelCli/utils/schema_model"
)

type DatabaseControllerInterface interface {
	dropTables() error
	dropEnums() error
	createTables(enumNames []string, modelNames []string, models []schema_model2.Model) error
	createEnums(enums []schema_model2.Enum) error
	RunMigrations(schema *schema_model2.GoRelSchema, enumNames []string, modelNames []string) error
	Close() error
	checkConnection() error
}
