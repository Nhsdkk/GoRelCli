package database_contoller

type DatabaseControllerInterface interface {
	dropTables() *DatabaseError
	dropEnums() *DatabaseError
	createTables() *DatabaseError
	createEnums() *DatabaseError
	RunMigrations() *DatabaseError
	Close() *DatabaseError
}
