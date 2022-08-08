package miniredis

import (
	"strconv"

	"github.com/alicebob/miniredis/v2/server"
)

// optInt parses an int option in a command.
// Writes "invalid integer" error to c if it's not a valid integer. Returns
// whether or not things were okay.
func optInt(c *server.Peer, src string, dest *int) bool {
	n, err := strconv.Atoi(src)
	if err != nil {
		setDirty(c)
		c.WriteError(msgInvalidInt)
		return false
	}
	*dest = n
	return true
}
