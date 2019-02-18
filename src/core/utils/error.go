package utils

import (
	"encoding/json"
	"errors"
)

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// WrapErrorMessage wraps the error msg to the well formated error message
// {
//		"error": {
//			"code": 404,
//			"message": "Requested entity was not found."
//		}
//}
func WrapErrorMessage(code int, message string) string {
	error := &Error{
		Code:    code,
		Message: message,
	}
	data, err := json.Marshal(&error)
	if err != nil {
		return message
	}
	return string(data)
}

// WrapError wraps the error to the well formated error `{ "error": "The error message" }`
func WrapError(code int, err error) error {
	if err == nil {
		return nil
	}

	return errors.New(WrapErrorMessage(code, err.Error()))
}
