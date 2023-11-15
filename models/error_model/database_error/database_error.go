package database_error

import "fmt"

type DatabaseErrorType string

const (
	UnsupportedProviderError DatabaseErrorType = "unsupported provider error"
	TransactionError                           = "transaction error"
	CloseConnectionError                       = "close connection error"
	ConnectionError                            = "connection error"
	SqlGenerationError                         = "sql generation error"
)

type DatabaseError struct {
	ErrorType DatabaseErrorType
	Text      string
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("Error while using database:\n%s - %s", e.ErrorType, e.Text)
}
