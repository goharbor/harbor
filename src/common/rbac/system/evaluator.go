package system

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/namespace"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// NewEvaluator create evaluator for the system
func NewEvaluator(username string, policies []*types.Policy) evaluator.Evaluator {
	return namespace.New(NamespaceKind, func(ctx context.Context, ns types.Namespace) evaluator.Evaluator {
		return rbac.New(&rbacUser{
			username: username,
			policies: policies,
		})
	})
}
