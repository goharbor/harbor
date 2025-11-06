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
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/utils/email"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/config"
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

	subject, ok := params["subject"]
	if !ok {
		return errors.Errorf("missing job parameter 'subject'")
	}
	_, ok = subject.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'subject', expecting string but got %s", reflect.TypeOf(subject).String())
	}

	body, ok := params["body"]
	if !ok {
		return errors.Errorf("missing job parameter 'body'")
	}
	_, ok = body.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'body', expecting string but got %s", reflect.TypeOf(body).String())
	}

	to, ok := params["to"]
	if !ok {
		return errors.Errorf("missing job parameter 'to'")
	}
	_, ok = to.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'to', expecting string but got %s", reflect.TypeOf(to).String())
	}
	return nil
}

// Run implements the interface in job/Interface
func (ej *EmailJob) Run(ctx job.Context, params job.Parameters) error {
	if err := ej.init(ctx, params); err != nil {
		return err
	}

	ej.logger.Info("start to run email job")

	err := ej.execute(ctx, params)
	if err != nil {
		ej.logger.Errorf("exit email job, error: %s", err)
	} else {
		ej.logger.Info("success to run email job")
	}
	return err
}

// init email job
func (ej *EmailJob) init(ctx job.Context, params map[string]any) error {
	ej.logger = ctx.GetLogger()
	return nil
}

// execute email job
func (ej *EmailJob) execute(ctx job.Context, params map[string]any) error {
	subject := params["subject"].(string)
	body := params["body"].(string)
	toStr := params["to"].(string)

	to := strings.Split(toStr, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}

	// Get email configuration from Harbor config
	cfg, err := config.Email(ctx.SystemContext())
	if err != nil {
		return errors.Wrap(err, "failed to get email config")
	}

	if cfg.Host == "" {
		return errors.New("email host not configured")
	}

	addr := cfg.Host
	if cfg.Port != 0 {
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}

	err = email.Send(
		addr,
		cfg.Identity,
		cfg.Username,
		cfg.Password,
		int(cfg.Timeout),
		cfg.SSL,
		cfg.Insecure,
		cfg.From,
		to,
		subject,
		body,
	)
	if err != nil {
		return errors.Wrap(err, "failed to send email")
	}
	return nil
}