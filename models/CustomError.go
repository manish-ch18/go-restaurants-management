package models

type CustomError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	Message     string `json:"message"`
}

func (e CustomError) GetError() string {
	return e.Message
}

func (e CustomError) GetCode() int {
	return e.Code
}

func (e CustomError) GetDescription() string {
	return e.Description
}

func ErrorResponse(code int, description, message string) CustomError {
	return CustomError{
		Code:        code,
		Description: description,
		Message:     message,
	}
}
