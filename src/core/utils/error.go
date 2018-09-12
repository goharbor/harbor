package utils

import (
	"encoding/json"
	"errors"
)

// WrapErrorMessage wraps the error msg to the well formated error message `{ "error": "The error message" }`
func WrapErrorMessage(msg string) string {
	errBody := make(map[string]string, 1)
	errBody["error"] = msg
	data, err := json.Marshal(&errBody)
	if err != nil {
		return msg
	}

	return string(data)
}

// WrapError wraps the error to the well formated error `{ "error": "The error message" }`
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(WrapErrorMessage(err.Error()))
}
