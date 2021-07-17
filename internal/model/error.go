package model

const (
	// 404 Error
	ErrNotFound = errorStr("not found")
	// 400 Error
	ErrBadRequest = errorStr("bad request")
)

// errorStr implements error interface
// and keeps primitive type's features (comparable, constants)
type errorStr string

// Error returns errorStr msg
// implements Error interface
func (err errorStr) Error() string {
	return string(err)
}
