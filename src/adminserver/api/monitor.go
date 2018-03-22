package api

import (
	"net/http"
	"github.com/vmware/harbor/src/common/utils/log"
)

// Ping monitor the server status
func Ping(w http.ResponseWriter, r *http.Request) {
	if err := writeJSON(w, "Pong"); err != nil {
		log.Errorf("Failed to write response: %v", err)
		return
	}
}
