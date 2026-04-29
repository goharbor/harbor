package defaults

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"github.com/volcengine/volcengine-go-sdk/internal/shareddefaults"
)

// SharedCredentialsFilename returns the SDK's default file path
// for the shared credentials file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.volcengine/credentials
//   - Windows: %USERPROFILE%\.volcengine\credentials
func SharedCredentialsFilename() string {
	return shareddefaults.SharedCredentialsFilename()
}

// SharedConfigFilename returns the SDK's default file path for
// the shared config file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.volcengine/config
//   - Windows: %USERPROFILE%\.volcengine\config
func SharedConfigFilename() string {
	return shareddefaults.SharedConfigFilename()
}

// SharedEndpointConfigFilename returns the SDK's default file path for
// the endpoint config file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.volcengine/endpoint
//   - Windows: %USERPROFILE%\.volcengine\endpoint
func SharedEndpointConfigFilename() string {
	return shareddefaults.SharedEndpointConfigFilename()
}
