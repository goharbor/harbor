// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notification

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/utils/email"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

// EmailJob implements the job interface, which send notification via email.
type EmailJob struct {
	logger logger.Interface
}

// MaxFails returns that how many times this job can fail.
func (ej *EmailJob) MaxFails() (result uint) {
	// Default max fails count is 3
	result = 3
	if maxFails, exist := os.LookupEnv(maxFails); exist {
		mf, err := strconv.ParseUint(maxFails, 10, 32)
		if err != nil {
			logger.Warningf("Fetch email job maxFails error: %s", err.Error())
			return result
		}
		result = uint(mf)
	}
	return result
}

// MaxCurrency is implementation of same method in Interface.
func (ej *EmailJob) MaxCurrency() uint {
	return 1
}

// ShouldRetry ...
func (ej *EmailJob) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (ej *EmailJob) Validate(params job.Parameters) error {
	if params == nil {
		// Params are required
		return errors.New("missing parameter of email job")
	}

	requiredParams := []string{"subject", "body", "to", "address", "from"}
	for _, param := range requiredParams {
		value, ok := params[param]
		if !ok {
			return errors.Errorf("missing job parameter '%s'", param)
		}
		_, ok = value.(string)
		if !ok {
			return errors.Errorf("malformed job parameter '%s', expecting string but got %s", param, reflect.TypeOf(value).String())
		}
	}
	return nil
}

// Run implements the interface in job/Interface
func (ej *EmailJob) Run(ctx job.Context, params job.Parameters) error {
	if err := ej.init(ctx, params); err != nil {
		return err
	}

	ej.logger.Info("start to run email job")

	err := ej.execute(params)
	if err != nil {
		ej.logger.Errorf("exit email job, error: %s", err)
	} else {
		ej.logger.Info("success to run email job")
	}
	return err
}

// init email job
func (ej *EmailJob) init(ctx job.Context, _ map[string]any) error {
	ej.logger = ctx.GetLogger()
	return nil
}

// execute email job
func (ej *EmailJob) execute(params map[string]any) error {
	subject := params["subject"].(string)
	body := params["body"].(string)
	toStr := params["to"].(string)
	address := params["address"].(string)
	from := params["from"].(string)

	username, _ := params["username"].(string)
	password, _ := params["password"].(string)
	identity, _ := params["identity"].(string)

	useSSL := true
	if ssl, ok := params["use_ssl"].(bool); ok {
		useSSL = ssl
	}

	insecure := false
	if insecureSkip, ok := params["insecure_skip_verify"].(bool); ok {
		insecure = insecureSkip
	}

	timeout := 60
	if tm, ok := params["timeout"].(int); ok {
		timeout = tm
	}

	port := 465
	if p, ok := params["port"].(int); ok {
		port = p
	}

	to := strings.Split(toStr, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}

	if address == "" {
		return errors.New("email address not configured")
	}

	if port != 0 && !strings.Contains(address, ":") {
		address = address + ":" + strconv.Itoa(port)
	}

	err := email.Send(
		address,
		identity,
		username,
		password,
		timeout,
		useSSL,
		insecure,
		from,
		to,
		subject,
		body,
	)
	if err != nil {
		return errors.Wrap(err, "failed to send email")
	}
	return nil
}
