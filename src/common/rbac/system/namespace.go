package system

import (
	"fmt"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"strings"
)

const (
	// NamespaceKind kind for system namespace
	NamespaceKind = "system"
	// NamespacePrefix for system namespace
	NamespacePrefix = "/system"
)

type systemNamespace struct {
}

func (ns *systemNamespace) Kind() string {
	return NamespaceKind
}

func (ns *systemNamespace) Resource(subresources ...types.Resource) types.Resource {
	return types.Resource(fmt.Sprintf("/system/")).Subresource(subresources...)
}

func (ns *systemNamespace) Identity() interface{} {
	return nil
}

func (ns *systemNamespace) GetPolicies() []*types.Policy {
	return policies
}

// NewNamespace returns namespace for project
func NewNamespace() types.Namespace {
	return &systemNamespace{}
}

// NamespaceParse ...
func NamespaceParse(resource types.Resource) (types.Namespace, bool) {
	if strings.HasPrefix(resource.String(), NamespacePrefix) {
		return NewNamespace(), true
	}
	return nil, false
}

func init() {
	types.RegistryNamespaceParse(NamespaceKind, NamespaceParse)
}
