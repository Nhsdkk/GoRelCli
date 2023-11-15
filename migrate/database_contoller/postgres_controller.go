package database_contoller

import (
	"GoRelCli/models/error_model/database_error"
	"GoRelCli/models/schema_model"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"slices"
)

type databaseEnum struct {
	Oid      int
	TypeName string
}

type PostgresController struct {
	db *sql.DB
}

func (p *PostgresController) dropTables() error {
	tableNames, err := p.getTables()

	if err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.TransactionError,
			Text:      fmt.Sprintf("Can't get table names: %s", err),
		}
	}

	if len(tableNames) == 0 {
		fmt.Println("No tables found. Skipping drop tables...")
		return nil
	}

	var queries []string
	for _, tableName := range tableNames {
		query := p.generateDeleteTableSqlScriptFromDbTableName(tableName)
		queries = append(queries, query)
	}
	rawSqlString := p.generateTransaction(queries)
	_, err = p.db.Exec(rawSqlString)

	if err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.TransactionError,
			Text:      fmt.Sprintf("Can't drop tables: %s", err),
		}
	}

	return nil
}

func (p *PostgresController) dropEnums() error {
	dbEnums, err := p.getEnums()

	if err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.TransactionError,
			Text:      fmt.Sprintf("Can't get enum names: %s", err),
		}
	}

	if len(dbEnums) == 0 {
		fmt.Println("No enums found. Skipping drop enums...")
		return nil
	}

	var queries []string
	for _, dbEnum := range dbEnums {
		queries = append(queries, p.generateDeleteEnumSqlScriptFromDbEnum(dbEnum))
	}

	rawSqlQuery := p.generateTransaction(queries)

	if _, err = p.db.Exec(rawSqlQuery); err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.TransactionError,
			Text:      fmt.Sprintf("Can't drop enums: %s", err),
		}
	}

	return nil
}

func (p *PostgresController) createTables(enumNames []string, modelNames []string, models []schema_model.Model) error {
	var createTableQueries []string
	var createRelationsQueries []string
	for _, model := range models {
		createTableQuery, createRelationsQueriesInner, err := p.generateCreateTableWithoutRelationsSqlScriptFromModel(model, enumNames, modelNames)
		if err != nil {
			return database_error.DatabaseError{
				ErrorType: database_error.SqlGenerationError,
				Text:      fmt.Sprintf("Can't generate sql query for relations and tables creation: %s", err),
			}
		}
		createTableQueries = append(createTableQueries, createTableQuery)
		createRelationsQueries = append(createRelationsQueries, createRelationsQueriesInner...)
	}

	createTablesRawSQLQuery := p.generateTransaction(createTableQueries)
	createRelationsRawSQLQuery := p.generateTransaction(createRelationsQueries)

	fmt.Println(fmt.Sprintf("Generated query for creating tables:\n%s", createTablesRawSQLQuery))
	fmt.Println(fmt.Sprintf("Generated query for creating relations:\n%s", createRelationsRawSQLQuery))

	if _, err := p.db.Exec(createTablesRawSQLQuery); err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.TransactionError,
			Text:      fmt.Sprintf("Can't create tables: %s", err),
		}
	}

	if _, err := p.db.Exec(createRelationsRawSQLQuery); err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.TransactionError,
			Text:      fmt.Sprintf("Can't create relations: %s", err),
		}
	}

	return nil
}

func (p *PostgresController) createEnums(enums []schema_model.Enum) error {
	var queries []string
	for _, enum := range enums {
		queries = append(queries, p.generateCreateEnumSqlScriptFromEnum(enum))
	}
	rawSqlQuery := p.generateTransaction(queries)
	fmt.Println(fmt.Sprintf("Generated query for creating enums:\n%s", rawSqlQuery))
	if _, err := p.db.Exec(rawSqlQuery); err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.TransactionError,
			Text:      fmt.Sprintf("Can't create enums: %s", err),
		}
	}
	return nil
}

func (p *PostgresController) RunMigrations(schema *schema_model.GoRelSchema, enumNames []string, modelNames []string) error {
	if err := p.dropTables(); err != nil {
		return err
	}
	if err := p.dropEnums(); err != nil {
		return err
	}
	if err := p.createEnums(schema.Enums); err != nil {
		return err
	}
	if err := p.createTables(enumNames, modelNames, schema.Models); err != nil {
		return err
	}
	return nil
}

func (p *PostgresController) Close() error {
	if err := p.db.Close(); err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.CloseConnectionError,
			Text:      fmt.Sprintf("Can't close connection to postgres db."),
		}
	}
	return nil
}

func (p *PostgresController) generateTransaction(sqlQueries []string) string {
	sqlString := "BEGIN;\n"
	for _, sqlQuery := range sqlQueries {
		sqlString = sqlString + sqlQuery + "\n"
	}
	sqlString = sqlString + "COMMIT;"
	return sqlString
}

func (p *PostgresController) getEnums() ([]databaseEnum, error) {
	/*
		select OID
		FROM pg_type
		WHERE OID = ANY (select enumtypid FROM pg_enum);
	*/
	const rawSqlString = "SELECT OID,TYPNAME FROM pg_type WHERE OID = ANY(SELECT enumtypid FROM pg_enum)"

	rows, err := p.db.Query(rawSqlString)
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

func (p *PostgresController) getTables() ([]string, error) {
	/*
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public';
	*/
	const rawSqlString = "select table_name from information_schema.tables where table_schema = 'public'"

	rows, err := p.db.Query(rawSqlString)
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

func (p *PostgresController) checkConnection() error {
	if err := p.db.Ping(); err != nil {
		return database_error.DatabaseError{
			ErrorType: database_error.ConnectionError,
			Text:      fmt.Sprintf("Can't connect to db: %s", err),
		}
	}
	return nil
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
//				return nil, error_model.New(fmt.Sprintf("can't find env variable with name %s", envVariableName))
//			}
//			db, err = getPostgreSQLController(urlEnv)
//		}
//	default:
//		return nil, error_model.New("provider not supported")
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

func (p *PostgresController) generateDeleteTableSqlScriptFromDbTableName(tableName string) string {
	//DROP TABLE User CASCADE;
	return fmt.Sprintf("DROP TABLE \"%s\" CASCADE;", tableName)
}

func (p *PostgresController) generateRelationsSqlScriptFromProperty(property schema_model.Property, modelName string) string {
	//ALTER TABLE "Todo" ADD CONSTRAINT "fk_User" FOREIGN KEY ("userId") REFERENCES "User" ("id");
	relationTableName := modelName
	referenceTableName := property.Type
	return fmt.Sprintf("ALTER TABLE \"%s\" ADD CONSTRAINT \"fk_%s\" FOREIGN KEY (\"%s\") REFERENCES \"%s\" (\"%s\");", relationTableName, referenceTableName, property.RelationField, referenceTableName, property.ReferenceField)
}

func (p *PostgresController) addTypeProperty(sqlQuery *string, property schema_model.Property, enumNames []string) error {
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

func (p *PostgresController) addIdProperty(sqlQuery *string, property schema_model.Property) {
	if property.Id {
		*sqlQuery += " PRIMARY KEY"
	}
}

func (p *PostgresController) addUniqueProperty(sqlQuery *string, property schema_model.Property) {
	if property.Unique {
		*sqlQuery += " UNIQUE"
	}
}

func (p *PostgresController) addDefaultProperty(sqlQuery *string, property schema_model.Property) error {
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

func (p *PostgresController) generateCreateTableWithoutRelationsSqlScriptFromModel(model schema_model.Model, enumNames []string, tableNames []string) (string, []string, error) {

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

			if err := p.addTypeProperty(&propertyString, property, enumNames); err != nil {
				return "", nil, err
			}
			p.addIdProperty(&propertyString, property)
			p.addUniqueProperty(&propertyString, property)
			if err := p.addDefaultProperty(&propertyString, property); err != nil {
				return "", nil, err
			}

			propertyString += ","

			rawSqlQuery = rawSqlQuery + propertyString
			continue
		}

		//fmt.Println("GENERATING RELATIONS FOR TABLE " + property.Type)
		relationString := p.generateRelationsSqlScriptFromProperty(property, model.Name)
		relationQueries = append(relationQueries, relationString)

	}

	rawSqlQuery = rawSqlQuery[:len(rawSqlQuery)-1] + ");"

	return rawSqlQuery, relationQueries, nil
}

func (p *PostgresController) generateDeleteEnumSqlScriptFromDbEnum(enum databaseEnum) string {
	//DROP TYPE UserRole;
	return fmt.Sprintf("DROP TYPE \"%s\";", enum.TypeName)
}

func (p *PostgresController) generateCreateEnumSqlScriptFromEnum(enum schema_model.Enum) string {
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
