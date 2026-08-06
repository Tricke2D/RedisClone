package server

import "fmt"

// RedisError represents Redis protocol errors
type RedisError struct {
	Code    string // Kode error (ERR, WRONGTYPE, dll)
	Message string
}

// Error implements error interface
func (e *RedisError) Error() string {
	return fmt.Sprintf("%s %s", e.Code, e.Message)
}

// ErrUnknownCommand: command tidak dikenal
func ErrUnknownCommand(cmd string) *RedisError {
	return &RedisError{Code: "ERR", Message: fmt.Sprintf("unknown command '%s'", cmd)}
}

// ErrWrongType: operasi terhadap key dengan tipe data yang salah
func ErrWrongType() *RedisError {
	return &RedisError{Code: "WRONGTYPE", Message: "Operation against a key holding the wrong kind of value"}
}

// ErrSyntax: syntax command salah
func ErrSyntax() *RedisError {
	return &RedisError{Code: "ERR", Message: "syntax error"}
}

// ErrWrongNumArgs: jumlah argumen command salah
func ErrWrongNumArgs(cmd string) *RedisError {
	return &RedisError{Code: "ERR", Message: fmt.Sprintf("wrong number of arguments for '%s' command", cmd)}
}