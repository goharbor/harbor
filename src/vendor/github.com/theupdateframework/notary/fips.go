package notary

import (
	"crypto"
	// Need to import md5 so can test availability.
	_ "crypto/md5"
)

// FIPSEnabled returns true if running in FIPS mode.
// If compiled in FIPS mode the md5 hash function is never available
// even when imported. This seems to be the best test we have for it.
func FIPSEnabled() bool {
	return !crypto.MD5.Available()
}
