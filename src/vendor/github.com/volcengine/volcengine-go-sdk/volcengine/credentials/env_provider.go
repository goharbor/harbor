package credentials

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"os"

	"github.com/volcengine/volcengine-go-sdk/volcengine/volcengineerr"
)

// EnvProviderName provides a name of Env provider
const EnvProviderName = "EnvProvider"

var (
	// ErrAccessKeyIDNotFound is returned when the Volcengine Access Key ID can't be
	// found in the process's environment.
	ErrAccessKeyIDNotFound = volcengineerr.New("EnvAccessKeyNotFound", "VOLCSTACK_ACCESS_KEY_ID or VOLCSTACK_ACCESS_KEY not found in environment", nil)

	// ErrSecretAccessKeyNotFound is returned when the Volcengine Secret Access Key
	// can't be found in the process's environment.
	ErrSecretAccessKeyNotFound = volcengineerr.New("EnvSecretNotFound", "VOLCSTACK_SECRET_ACCESS_KEY or VOLCSTACK_SECRET_KEY not found in environment", nil)
)

// A EnvProvider retrieves credentials from the environment variables of the
// running process. Environment credentials never expire.
//
// Environment variables used:
//
// * Access Key ID:     VOLCSTACK_ACCESS_KEY_ID or VOLCSTACK_ACCESS_KEY
//
// * Secret Access Key: VOLCSTACK_SECRET_ACCESS_KEY or VOLCSTACK_SECRET_KEY
type EnvProvider struct {
	retrieved bool
}

// NewEnvCredentials returns a pointer to a new Credentials object
// wrapping the environment variable provider.
func NewEnvCredentials() *Credentials {
	return NewCredentials(&EnvProvider{})
}

// Retrieve retrieves the keys from the environment.
func (e *EnvProvider) Retrieve() (Value, error) {
	e.retrieved = false

	id := os.Getenv("VOLCSTACK_ACCESS_KEY_ID")
	if id == "" {
		id = os.Getenv("VOLCSTACK_ACCESS_KEY")
	}

	secret := os.Getenv("VOLCSTACK_SECRET_ACCESS_KEY")
	if secret == "" {
		secret = os.Getenv("VOLCSTACK_SECRET_KEY")
	}

	if id == "" {
		return Value{ProviderName: EnvProviderName}, ErrAccessKeyIDNotFound
	}

	if secret == "" {
		return Value{ProviderName: EnvProviderName}, ErrSecretAccessKeyNotFound
	}

	e.retrieved = true
	return Value{
		AccessKeyID:     id,
		SecretAccessKey: secret,
		SessionToken:    os.Getenv("VOLCSTACK_SESSION_TOKEN"),
		ProviderName:    EnvProviderName,
	}, nil
}

// IsExpired returns if the credentials have been retrieved.
func (e *EnvProvider) IsExpired() bool {
	return !e.retrieved
}
