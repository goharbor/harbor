package webhook

import (
	"bytes"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"net/http"
)

// HTTPNotifier implements the job interface, which send webhook notification by http.
type HTTPNotifier struct {
	client *http.Client
	logger logger.Interface
	ctx    job.Context
}

// MaxFails returns that how many times this job can fail, get this value from ctx.
func (hn *HTTPNotifier) MaxFails() uint {
	// Get max fails count from config file
	// Default max fails count is 10, and its max retry interval is around 3h
	// Large enough to ensure most situations can notify successfully
	return config.DefaultConfig.WebHookConfig.MaxHttpFails
}

// ShouldRetry ...
func (hn *HTTPNotifier) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (hn *HTTPNotifier) Validate(params job.Parameters) error {
	return nil
}

// Run implements the interface in job/Interface
func (hn *HTTPNotifier) Run(ctx job.Context, params job.Parameters) error {
	if err := hn.init(ctx, params); err != nil {
		return err
	}

	err := hn.execute(ctx, params)
	return err
}

// init http_notifier for webhoook
func (hn *HTTPNotifier) init(ctx job.Context, params map[string]interface{}) error {
	hn.logger = ctx.GetLogger()
	hn.ctx = ctx

	// default insecureSkipVerify is false
	insecureSkipVerify := false
	if v, ok := params["skip_cert_verify"]; ok {
		insecureSkipVerify = v.(bool)
	}
	hn.client = &http.Client{
		Transport: registry.GetHTTPTransport(insecureSkipVerify),
	}

	return nil
}

// send notification by http or https
func (hn *HTTPNotifier) execute(ctx job.Context, params map[string]interface{}) error {
	payload := params["payload"].(string)
	address := params["address"].(string)

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader([]byte(payload)))
	if err != nil {
		return err
	}
	if v, ok := params["secret"]; ok && len(v.(string)) > 0 {
		req.Header.Set("Authorization", "Secret "+v.(string))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := hn.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		return fmt.Errorf("webhook job(target: %s) response code is %d", address, resp.StatusCode)
	}

	return nil
}
