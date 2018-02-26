package utils

import (
	"net/http"

	// have to use this context version to satisfy AccessController interface :-/
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/auth"
)

// TestingAccessController is for TEST USE ONLY!!!
// It allows you to configure an AccessController that always
// succeeds or fails.
type TestingAccessController struct {
	Err error
}

var _ auth.AccessController = TestingAccessController{}

// Authorized will either always error, or always succeed, depending on the
// configuration of the TestingAccessController
func (ac TestingAccessController) Authorized(ctx context.Context, access ...auth.Access) (context.Context, error) {
	return ctx, ac.Err
}

// TestingAuthChallenge is for TEST USE ONLY!!!
// It implements the auth.Challenge interface and allows a test to confirm
// the the SetHeaders method was called.
type TestingAuthChallenge struct {
	SetHeadersCalled bool
}

var _ auth.Challenge = &TestingAuthChallenge{}

func (c TestingAuthChallenge) Error() string {
	return "TestingAuthChallenge"
}

// SetHeaders just records that the function was called for the purpose of testing
func (c *TestingAuthChallenge) SetHeaders(w http.ResponseWriter) {
	c.SetHeadersCalled = true
}
