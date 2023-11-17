package gdpr

import (
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/pkg/user"
)

const UserNameParam = "username"

type AuditLogsDataMasking struct {
	manager     audit.Manager
	userManager user.Manager
}

func (a AuditLogsDataMasking) MaxFails() uint {
	return 3
}

func (a AuditLogsDataMasking) MaxCurrency() uint {
	return 1
}

func (a AuditLogsDataMasking) ShouldRetry() bool {
	return true
}

func (a AuditLogsDataMasking) Validate(params job.Parameters) error {
	if params == nil {
		// Params are required
		return errors.New("missing job parameters")
	}
	_, err := a.parseParams(params)
	return err
}

func (a *AuditLogsDataMasking) init() {
	if a.manager == nil {
		a.manager = audit.New()
	}
	if a.userManager == nil {
		a.userManager = user.New()
	}
}

func (a AuditLogsDataMasking) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()
	logger.Info("GDPR audit logs data masking job started")
	a.init()
	username, err := a.parseParams(params)
	if err != nil {
		return err
	}
	logger.Infof("Masking log entries for a user: %s", username)
	return a.manager.UpdateUsername(ctx.SystemContext(), username, a.userManager.GenerateCheckSum(username))
}

func (a AuditLogsDataMasking) parseParams(params job.Parameters) (string, error) {
	value, exist := params[UserNameParam]
	if !exist {
		return "", fmt.Errorf("param %s not found", UserNameParam)
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("the value of %s isn't string", UserNameParam)
	}
	return str, nil
}
