package database_contoller

import (
	"GoRelCli/schema_model"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"slices"
)

type DatabaseErrorType string

const (
	UnsupportedProviderError DatabaseErrorType = "unsupported provider error"
	TransactionError                           = "transaction error"
	CloseConnectionError                       = "close connection error"
	ConnectionError                            = "connection error"
)

type databaseEnum struct {
	Oid      int
	TypeName string
}

type DatabaseError struct {
	errorType DatabaseErrorType
	text      string
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("Error while using database: %s - %s", e.errorType, e.text)
}

type PostgresController struct {
	db *sql.DB
}

func (p *PostgresController) dropTables() *DatabaseError {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresController) dropEnums() *DatabaseError {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresController) createTables() *DatabaseError {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresController) createEnums() *DatabaseError {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresController) RunMigrations() *DatabaseError {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresController) Close() *DatabaseError {
	if err := p.db.Close(); err != nil {
		return &DatabaseError{
			errorType: CloseConnectionError,
			text:      fmt.Sprintf("Error while closing connection to postgres db."),
		}
	}
	return nil
}

func generateTransaction(sqlQueries []string) string {
	sqlString := "BEGIN;\n"
	for _, sqlQuery := range sqlQueries {
		sqlString = sqlString + sqlQuery + "\n"
	}
	sqlString = sqlString + "COMMIT;"
	return sqlString
}

func getEnums(db *sql.DB) ([]databaseEnum, error) {
	/*
		select OID
		FROM pg_type
		WHERE OID = ANY (select enumtypid FROM pg_enum);
	*/
	const rawSqlString = "SELECT OID,TYPNAME FROM pg_type WHERE OID = ANY(SELECT enumtypid FROM pg_enum)"

	rows, err := db.Query(rawSqlString)
	if err != nil {
		return nil, err
	}

	var enums []databaseEnum
	for rows.Next() {
		enum := databaseEnum{}
		if err := rows.Scan(&enum.Oid, &enum.TypeName); err != nil {
			return nil, err
		}
		enums = append(enums, enum)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return enums, nil
}

func getTables(db *sql.DB) ([]string, error) {
	//SELECT table_name
	//FROM information_schema.tables
	//WHERE table_schema = 'public';
	const rawSqlString = "select table_name from information_schema.tables where table_schema = 'public'"

	rows, err := db.Query(rawSqlString)
	if err != nil {
		return nil, err
	}

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func checkConnection(db *sql.DB) error {
	return db.Ping()
}

//func GetController(provider schema_model.Provider, url string) (*sql.DB, error) {
//	var err error
//	var db *sql.DB
//	switch provider {
//	case schema_model.PostgreSQL:
//		isEnvFunc, err := regexp.MatchString("^env\\(\\\"\\S*\\\"\\)$", url)
//		if err != nil || !isEnvFunc {
//			db, err = getPostgreSQLController(url)
//		} else {
//			if err := env_loader.LoadEnvFile(); err != nil {
//				return nil, err
//			}
//			envVariableName := url[5 : len(url)-2]
//			urlEnv, exists := os.LookupEnv(envVariableName)
//			if !exists {
//				return nil, errors.New(fmt.Sprintf("can't find env variable with name %s", envVariableName))
//			}
//			db, err = getPostgreSQLController(urlEnv)
//		}
//	default:
//		return nil, errors.New("provider not supported")
//	}
//
//	if err != nil {
//		return nil, err
//	}
//
//	if err := checkConnection(db); err != nil {
//		return nil, err
//	}
//	return db, nil
//
//}

func generateDeleteTableSqlScriptFromDbTableName(tableName string) string {
	//DROP TABLE User CASCADE;
	return fmt.Sprintf("DROP TABLE \"%s\" CASCADE;", tableName)
}

func DropTables(db *sql.DB) error {
	tableNames, err := getTables(db)

	if err != nil {
		return err
	}

	if len(tableNames) == 0 {
		fmt.Println("No tables found. Skipping drop tables...")
		return nil
	}

	var queries []string
	for _, tableName := range tableNames {
		query := generateDeleteTableSqlScriptFromDbTableName(tableName)
		queries = append(queries, query)
	}
	rawSqlString := generateTransaction(queries)
	_, err = db.Exec(rawSqlString)
	return err
}

func generateRelationsSqlScriptFromProperty(property schema_model.Property, modelName string) string {
	//ALTER TABLE "Todo" ADD CONSTRAINT "fk_User" FOREIGN KEY ("userId") REFERENCES "User" ("id");
	relationTableName := modelName
	referenceTableName := property.Type
	return fmt.Sprintf("ALTER TABLE \"%s\" ADD CONSTRAINT \"fk_%s\" FOREIGN KEY (\"%s\") REFERENCES \"%s\" (\"%s\");", relationTableName, referenceTableName, property.RelationField, referenceTableName, property.ReferenceField)
}

func addTypeProperty(sqlQuery *string, property schema_model.Property, enumNames []string) error {
	var postgresType string
	isEnum := slices.Contains(enumNames, property.Type) || (slices.Contains(enumNames, property.Type[0:len(property.Type)-1]) && property.Type[len(property.Type)-1:len(property.Type)] == "?")

	if !isEnum {
		pType, isValidType := property.GetPostgresType()
		if !isValidType {
			return errors.New(fmt.Sprintf("Invalid property type provided: %s", property.Type))
		}
		postgresType = pType
	} else {
		postgresType = property.Type
	}

	if property.Default == "autoincrement()" {
		*sqlQuery += " SERIAL"
		return nil
	}

	if property.Default == "uuid()" {
		*sqlQuery += "  uuid"
		return nil
	}

	isOptional := property.Type[len(property.Type)-1:len(property.Type)] == "?"

	if isEnum && isOptional {
		*sqlQuery += fmt.Sprintf(" \"%s\"", postgresType[0:len(postgresType)-1])
	} else if isEnum && !isOptional {
		*sqlQuery += fmt.Sprintf(" \"%s\" NOT NULL", postgresType)
	} else {
		*sqlQuery += fmt.Sprintf(" %s", postgresType)
	}

	return nil
}

func addIdProperty(sqlQuery *string, property schema_model.Property) {
	if property.Id {
		*sqlQuery += " PRIMARY KEY"
	}
}

func addUniqueProperty(sqlQuery *string, property schema_model.Property) {
	if property.Unique {
		*sqlQuery += " UNIQUE"
	}
}

func addDefaultProperty(sqlQuery *string, property schema_model.Property) error {
	if property.Default == "uuid()" {
		*sqlQuery += "  DEFAULT(gen_random_uuid())"
		return nil
	}
	if property.Default == "now()" {
		*sqlQuery += "  DEFAULT(now())"
		return nil
	}
	if property.Default != "" && property.Default != "autoincrement()" {
		//fmt.Println(fmt.Sprintf("Property %s has default value %s of type %T", property.Name, property.Default, property.Default))
		_, err := property.ValidateDefaultValue()
		if err != nil {
			return err
		}
		*sqlQuery += fmt.Sprintf(" DEFAULT(%s)", property.Default)
	}
	return nil
}

func generateCreateTableWithoutRelationsSqlScriptFromModel(model schema_model.Model, enumNames []string, tableNames []string) (string, []string, error) {

	for index, tableName := range tableNames {
		tableNames[index] = tableName + "[]"
	}

	var relationQueries []string
	rawSqlQuery := fmt.Sprintf("CREATE TABLE \"%s\" (", model.Name)
	for _, property := range model.Properties {
		if slices.Contains(tableNames, property.Type) {
			continue
		}

		if property.RelationField == "" && property.ReferenceField == "" {
			var propertyString string

			propertyString = fmt.Sprintf("\"%s\"", property.Name)

			if err := addTypeProperty(&propertyString, property, enumNames); err != nil {
				return "", nil, err
			}
			addIdProperty(&propertyString, property)
			addUniqueProperty(&propertyString, property)
			if err := addDefaultProperty(&propertyString, property); err != nil {
				return "", nil, err
			}

			propertyString += ","

			rawSqlQuery = rawSqlQuery + propertyString
			continue
		}

		//fmt.Println("GENERATING RELATIONS FOR TABLE " + property.Type)
		relationString := generateRelationsSqlScriptFromProperty(property, model.Name)
		relationQueries = append(relationQueries, relationString)

	}

	rawSqlQuery = rawSqlQuery[:len(rawSqlQuery)-1] + ");"

	return rawSqlQuery, relationQueries, nil
}

func generateDeleteEnumSqlScriptFromDbEnum(enum databaseEnum) string {
	//DROP TYPE UserRole;
	return fmt.Sprintf("DROP TYPE \"%s\";", enum.TypeName)
}

func generateCreateEnumSqlScriptFromEnum(enum schema_model.Enum) string {
	//CREATE TYPE UserRole AS ENUM('Admin','User');
	rawSqlString := fmt.Sprintf("CREATE TYPE \"%s\" AS ENUM (", enum.Name)
	for index, value := range enum.Values {
		if index == len(enum.Values)-1 {
			rawSqlString += fmt.Sprintf("'%s');", value)
			continue
		}
		rawSqlString += fmt.Sprintf("'%s',", value)
	}
	return rawSqlString
}

func DropEnums(db *sql.DB) error {
	dbEnums, err := getEnums(db)

	if err != nil {
		return err
	}

	if len(dbEnums) == 0 {
		fmt.Println("No enums found. Skipping drop enums...")
		return nil
	}

	var queries []string
	for _, dbEnum := range dbEnums {
		queries = append(queries, generateDeleteEnumSqlScriptFromDbEnum(dbEnum))
	}

	rawSqlQuery := generateTransaction(queries)

	_, err = db.Exec(rawSqlQuery)
	return err
}

func CreateEnums(db *sql.DB, enums []schema_model.Enum) error {
	var queries []string
	for _, enum := range enums {
		queries = append(queries, generateCreateEnumSqlScriptFromEnum(enum))
	}
	rawSqlQuery := generateTransaction(queries)
	//fmt.Println(rawSqlQuery)
	_, err := db.Exec(rawSqlQuery)

	return err

}

func CreateTables(db *sql.DB, models []schema_model.Model, enumNames []string, modelNames []string) error {
	var createTableQueries []string
	var createRelationsQueries []string
	for _, model := range models {
		createTableQuery, createRelationsQueriesInner, err := generateCreateTableWithoutRelationsSqlScriptFromModel(model, enumNames, modelNames)
		if err != nil {
			return err
		}
		createTableQueries = append(createTableQueries, createTableQuery)
		createRelationsQueries = append(createRelationsQueries, createRelationsQueriesInner...)
	}

	createTablesRawSQLQuery := generateTransaction(createTableQueries)
	createRelationsRawSQLQuery := generateTransaction(createRelationsQueries)

	if _, err := db.Exec(createTablesRawSQLQuery); err != nil {
		return errors.New(fmt.Sprintf("error while creating tables:\n%s", err))
	}

	if _, err := db.Exec(createRelationsRawSQLQuery); err != nil {
		return errors.New(fmt.Sprintf("error while creating relations:\n%s", err))
	}

	return nil
}
