package models

type Success struct {
	IsSuccess bool        `json:"success"`
	Msg       string      `json:"message"`
	Data      interface{} `json:"data"`
}

// IsSuccessful method returns whether the operation was successful
func (s Success) IsSuccessful() bool {
	return s.IsSuccess
}

// Message method returns a success message
func (s Success) GetMessage() string {
	return s.Msg
}

// Data method returns additional data related to the success
func (s Success) GetData() interface{} {
	return s.Data
}

// SuccessResponse is a constructor function to create a new Success instance
func SuccessResponse(success bool, message string, data interface{}) Success {
	return Success{
		IsSuccess: success,
		Msg:       message,
		Data:      data,
	}
}
