package notification

import (
	"net/http"
	"os"
	"strconv"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/jobservice/logger"
)

const (
	secure   = "secure"
	insecure = "insecure"

	// Max retry has the same meaning as max fails.
	maxFails = "JOBSERVICE_WEBHOOK_JOB_MAX_RETRY"
	// http client timeout for webhook job(seconds).
	httpClientTimeout = "JOBSERVICE_WEBHOOK_JOB_HTTP_CLIENT_TIMEOUT"
)

var (
	// timeout records the timeout for http client
	timeout    time.Duration
	httpHelper *HTTPHelper
)

func init() {
	// default timeout is 3 seconds
	timeout = 3 * time.Second
	if envTimeout, exist := os.LookupEnv(httpClientTimeout); exist {
		t, err := strconv.ParseInt(envTimeout, 10, 64)
		if err != nil {
			logger.Warningf("Failed to parse timeout from environment, error: %v", err)
			return
		}

		timeout = time.Duration(t) * time.Second
		logger.Debugf("Set the http client timeout to %v for webhook job", timeout)
	}
}

// HTTPHelper in charge of sending notification messages to remote endpoint
type HTTPHelper struct {
	clients map[string]*http.Client
}

func init() {
	httpHelper = &HTTPHelper{
		clients: map[string]*http.Client{},
	}
	httpHelper.clients[secure] = &http.Client{
		Transport: commonhttp.GetHTTPTransport(),
		Timeout:   timeout,
	}
	httpHelper.clients[insecure] = &http.Client{
		Transport: commonhttp.GetHTTPTransport(commonhttp.WithInsecure(true)),
		Timeout:   timeout,
	}
}
