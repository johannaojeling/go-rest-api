package models

import "fmt"

type ErrorMessage struct {
	Details string `json:"details"`
}

func NewErrorMessage(message string, a ...any) ErrorMessage {
	return ErrorMessage{
		Details: fmt.Sprintf(message, a...),
	}
}
