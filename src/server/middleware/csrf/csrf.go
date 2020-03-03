package csrf

import (
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	ierror "github.com/goharbor/harbor/src/internal/error"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/gorilla/csrf"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	csrfKeyEnv  = "CSRF_KEY"
	tokenHeader = "X-Harbor-CSRF-Token"
	tokenCookie = "__csrf"
)

var (
	once    sync.Once
	protect func(handler http.Handler) http.Handler
)

// attachToken makes sure if csrf generate a new token it will be included in the response header
func attachToken(w http.ResponseWriter, r *http.Request) {
	if t := csrf.Token(r); len(t) > 0 {
		http.SetCookie(w, &http.Cookie{
			Name:     tokenCookie,
			Secure:   true,
			Value:    t,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		log.Warningf("token not found in context, skip attaching")
	}
}

func handleError(w http.ResponseWriter, r *http.Request) {
	attachToken(w, r)
	serror.SendError(w, ierror.New(csrf.FailureReason(r)).WithCode(ierror.ForbiddenCode))
	return
}

func attach(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		attachToken(rw, req)
		handler.ServeHTTP(rw, req)
	})
}

// Middleware initialize the middleware to apply csrf selectively
func Middleware() func(handler http.Handler) http.Handler {
	once.Do(func() {
		key := os.Getenv(csrfKeyEnv)
		if len(key) != 32 {
			log.Warningf("Invalid CSRF key from environment: %s, generating random key...", key)
			key = utils.GenerateRandomString()

		}
		protect = csrf.Protect([]byte(key), csrf.RequestHeader(tokenHeader),
			csrf.ErrorHandler(http.HandlerFunc(handleError)),
			csrf.SameSite(csrf.SameSiteStrictMode),
			csrf.Path("/"))
	})
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		protect(attach(next)).ServeHTTP(rw, req)
	}, csrfSkipper)
}

// csrfSkipper makes sure only some of the uris accessed by non-UI client can skip the csrf check
func csrfSkipper(req *http.Request) bool {
	path := req.URL.Path
	// We can check the cookie directly b/c the filter and controllerRegistry is executed after middleware, so no session
	// cookie is added by beego.
	_, err := req.Cookie(config.SessionCookieName)
	hasSession := err == nil
	if (strings.HasPrefix(path, "/v2/") ||
		strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/chartrepo/") ||
		strings.HasPrefix(path, "/service/")) && !hasSession {
		return true
	}
	return false
}
