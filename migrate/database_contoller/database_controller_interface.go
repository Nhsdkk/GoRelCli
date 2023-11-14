package database_contoller

import "GoRelCli/schema_model"

type DatabaseControllerInterface interface {
	dropTables() error
	dropEnums() error
	createTables(enumNames []string, modelNames []string, models []schema_model.Model) error
	createEnums(enums []schema_model.Enum) error
	RunMigrations(schema *schema_model.GoRelSchema, enumNames []string, modelNames []string) error
	Close() error
	checkConnection() error
}
