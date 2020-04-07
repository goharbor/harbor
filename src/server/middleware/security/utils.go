package security

import (
	"net/http"
	"strings"
)

func bearerToken(req *http.Request) string {
	if req == nil {
		return ""
	}
	h := req.Header.Get("Authorization")
	token := strings.Split(h, "Bearer")
	if len(token) < 2 {
		return ""
	}
	return strings.TrimSpace(token[1])
}
