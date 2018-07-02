package notary

import "os"

// FIPSEnvVar is the name of the environment variable that is being used to switch
// between FIPS and non-FIPS mode
const FIPSEnvVar = "GOFIPS"

// FIPSEnabled returns true if environment variable `GOFIPS` has been set to enable
// FIPS mode
func FIPSEnabled() bool {
	return os.Getenv(FIPSEnvVar) != ""
}
