package utils

import "fmt"

type HTTPError struct {
	Code    int
	Message string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func NewHTTPError(code int, message string) HTTPError {
	return HTTPError{code, message}
}
