package apilib

import ()

//Password for go test
type Password struct {

	// The user's existing password.
	OldPassword string `json:"old_password,omitempty"`

	// New password for marking as to be updated.
	NewPassword string `json:"new_password,omitempty"`
}
