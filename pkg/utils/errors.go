package utils

type AppError struct {
	Status  int
	Code    string
	Message string
}

func (e *AppError) Error() string { return e.Message }

func ErrBadRequest(msg string) *AppError {
	return &AppError{Status: 400, Code: "INVALID_REQUEST", Message: msg}
}

func ErrNotFound(code, msg string) *AppError {
	return &AppError{Status: 404, Code: code, Message: msg}
}

func ErrUnprocessable(code, msg string) *AppError {
	return &AppError{Status: 422, Code: code, Message: msg}
}

func ErrInternal(msg string) *AppError {
	return &AppError{Status: 500, Code: "INTERNAL_ERROR", Message: msg}
}
