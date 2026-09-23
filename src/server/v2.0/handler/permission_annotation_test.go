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

package handler

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/rbac"
)

// swaggerPath is the API definition, relative to this package.
const swaggerPath = "../../../../api/v2.0/swagger.yaml"

// permissionDocPrefix marks the generated sentence inside an operation description.
const permissionDocPrefix = "Permissions required: "

// requiredPermission is the permission an operation checks before doing its work.
type requiredPermission struct {
	scope    string // "system-level" or "project-level"
	resource rbac.Resource
	action   rbac.Action
}

// docLine renders the sentence expected in the swagger description of the operation.
func (p requiredPermission) docLine() string {
	return fmt.Sprintf("%s%s `%s:%s`", permissionDocPrefix, p.scope, p.resource, p.action)
}

// swaggerOperation is the subset of a swagger operation this test looks at.
type swaggerOperation struct {
	OperationID string `yaml:"operationId"`
	Description string `yaml:"description"`
}

// swaggerOperations returns the operations of the API definition keyed by operation ID.
func swaggerOperations(t *testing.T) map[string]swaggerOperation {
	content, err := os.ReadFile(swaggerPath)
	require.NoError(t, err, "cannot read %s", swaggerPath)

	spec := struct {
		Paths map[string]map[string]swaggerOperation `yaml:"paths"`
	}{}
	require.NoError(t, yaml.Unmarshal(content, &spec))

	operations := map[string]swaggerOperation{}
	for path, methods := range spec.Paths {
		for method, operation := range methods {
			if operation.OperationID == "" {
				continue
			}
			_, exist := operations[operation.OperationID]
			require.False(t, exist, "duplicated operationId %q at %s %s", operation.OperationID, strings.ToUpper(method), path)
			operations[operation.OperationID] = operation
		}
	}
	require.NotEmpty(t, operations)
	return operations
}

// operationKey makes an operation ID comparable with the name of the handler method that
// serves it, the two only differ in casing, e.g. "pingOIDC" is served by PingOIDC.
func operationKey(name string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(name))
}

// rbacAction and rbacResource turn the name of a constant of the rbac package back into its
// value. They are written out instead of being reflected so that a constant which is added to
// src/common/rbac/const.go but not here makes this test fail instead of silently skipping an
// operation.
var (
	rbacActions = map[string]rbac.Action{
		"ActionAll": rbac.ActionAll, "ActionPull": rbac.ActionPull, "ActionPush": rbac.ActionPush,
		"ActionCreate": rbac.ActionCreate, "ActionRead": rbac.ActionRead, "ActionUpdate": rbac.ActionUpdate,
		"ActionDelete": rbac.ActionDelete, "ActionList": rbac.ActionList, "ActionOperate": rbac.ActionOperate,
		"ActionScannerPull": rbac.ActionScannerPull, "ActionStop": rbac.ActionStop,
	}
	rbacResources = map[string]rbac.Resource{
		"ResourceAll": rbac.ResourceAll, "ResourceConfiguration": rbac.ResourceConfiguration,
		"ResourceLabel": rbac.ResourceLabel, "ResourceLog": rbac.ResourceLog,
		"ResourceLdapUser": rbac.ResourceLdapUser, "ResourceMember": rbac.ResourceMember,
		"ResourceMetadata": rbac.ResourceMetadata, "ResourceQuota": rbac.ResourceQuota,
		"ResourceRepository": rbac.ResourceRepository, "ResourceTagRetention": rbac.ResourceTagRetention,
		"ResourceImmutableTag": rbac.ResourceImmutableTag, "ResourceRobot": rbac.ResourceRobot,
		"ResourceNotificationPolicy": rbac.ResourceNotificationPolicy, "ResourceScan": rbac.ResourceScan,
		"ResourceSBOM": rbac.ResourceSBOM, "ResourceScanner": rbac.ResourceScanner,
		"ResourceArtifact": rbac.ResourceArtifact, "ResourceTag": rbac.ResourceTag,
		"ResourceAccessory": rbac.ResourceAccessory, "ResourceArtifactAddition": rbac.ResourceArtifactAddition,
		"ResourceArtifactLabel": rbac.ResourceArtifactLabel, "ResourcePreatPolicy": rbac.ResourcePreatPolicy,
		"ResourcePreatInstance": rbac.ResourcePreatInstance, "ResourceSelf": rbac.ResourceSelf,
		"ResourceAuditLog": rbac.ResourceAuditLog, "ResourceCatalog": rbac.ResourceCatalog,
		"ResourceProject": rbac.ResourceProject, "ResourceUser": rbac.ResourceUser,
		"ResourceUserGroup": rbac.ResourceUserGroup, "ResourceRegistry": rbac.ResourceRegistry,
		"ResourceReplication": rbac.ResourceReplication, "ResourceDistribution": rbac.ResourceDistribution,
		"ResourceGarbageCollection": rbac.ResourceGarbageCollection, "ResourceScanAll": rbac.ResourceScanAll,
		"ResourceReplicationAdapter": rbac.ResourceReplicationAdapter,
		"ResourceReplicationPolicy":  rbac.ResourceReplicationPolicy,
		"ResourceSystemVolumes":      rbac.ResourceSystemVolumes, "ResourcePurgeAuditLog": rbac.ResourcePurgeAuditLog,
		"ResourceExportCVE": rbac.ResourceExportCVE, "ResourceJobServiceMonitor": rbac.ResourceJobServiceMonitor,
		"ResourceSecurityHub": rbac.ResourceSecurityHub,
	}
)

// rbacConstant returns the name of the rbac package constant an expression refers to, it
// reports false for anything that is not a plain "rbac.XXX" with the given prefix.
func rbacConstant(expression ast.Expr, prefix string) (string, bool) {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "rbac" || !strings.HasPrefix(selector.Sel.Name, prefix) {
		return "", false
	}
	return selector.Sel.Name, true
}

// permissionOf returns the permission a RequireSystemAccess or RequireProjectAccess call asks
// for. Only a call carrying exactly one literal action and one literal subresource is
// understood, so that an operation is left undocumented rather than documented wrongly.
func permissionOf(t *testing.T, call *ast.CallExpr) (requiredPermission, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return requiredPermission{}, false
	}

	var scope string
	var arguments []ast.Expr
	switch selector.Sel.Name {
	case "RequireSystemAccess": // ctx, action, subresource...
		if len(call.Args) < 1 {
			return requiredPermission{}, false
		}
		scope, arguments = "system-level", call.Args[1:]
	case "RequireProjectAccess": // ctx, projectIDOrName, action, subresource...
		if len(call.Args) < 2 {
			return requiredPermission{}, false
		}
		scope, arguments = "project-level", call.Args[2:]
	default:
		return requiredPermission{}, false
	}
	// Zero subresources means the project itself and more than one means a choice, the
	// sentence can express neither.
	if len(arguments) != 2 {
		return requiredPermission{}, false
	}

	actionName, ok := rbacConstant(arguments[0], "Action")
	if !ok {
		return requiredPermission{}, false
	}
	resourceName, ok := rbacConstant(arguments[1], "Resource")
	if !ok {
		return requiredPermission{}, false
	}
	action, ok := rbacActions[actionName]
	require.True(t, ok, "rbac.%s is unknown to this test, add it to rbacActions", actionName)
	resource, ok := rbacResources[resourceName]
	require.True(t, ok, "rbac.%s is unknown to this test, add it to rbacResources", resourceName)

	return requiredPermission{scope: scope, resource: resource, action: action}, true
}

// handlerMethod is a method of this package that serves a swagger operation.
type handlerMethod struct {
	name       string
	permission requiredPermission
	checked    bool // whether permission holds the permission the method requires of every request
}

// handlerMethods parses the handlers of this package and returns the methods serving the
// wanted operations, keyed the same way so that they can be looked up by operation ID. Methods
// that serve no operation, Prepare and the helpers of BaseAPI, are left out.
func handlerMethods(t *testing.T, wanted map[string]bool) map[string]handlerMethod {
	fileSet := token.NewFileSet()
	packages, err := parser.ParseDir(fileSet, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	require.NoError(t, err)
	require.Contains(t, packages, "handler")

	methods := map[string]handlerMethod{}
	for _, file := range packages["handler"].Files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil || function.Body == nil || !ast.IsExported(function.Name.Name) {
				continue
			}
			if !returnsResponder(function) {
				continue
			}
			key := operationKey(function.Name.Name)
			if !wanted[key] {
				continue
			}
			_, duplicated := methods[key]
			require.False(t, duplicated, "two handler methods serve the operation %q", function.Name.Name)
			methods[key] = handlerMethod{name: function.Name.Name}

			// Only the statements at the top level of the body are inspected, a check nested
			// in an if, for or switch does not apply to every request of the operation.
			for _, statement := range function.Body.List {
				ifStatement, ok := statement.(*ast.IfStmt)
				if !ok || ifStatement.Init == nil {
					continue
				}
				assignment, ok := ifStatement.Init.(*ast.AssignStmt)
				if !ok || len(assignment.Rhs) != 1 {
					continue
				}
				call, ok := assignment.Rhs[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				permission, ok := permissionOf(t, call)
				if !ok {
					continue
				}
				method := methods[key]
				if method.checked {
					// A second check, one sentence would only tell half the story.
					method.checked = false
					methods[key] = method
					break
				}
				methods[key] = handlerMethod{name: function.Name.Name, permission: permission, checked: true}
			}
		}
	}
	require.NotEmpty(t, methods)
	return methods
}

// returnsResponder reports whether a function returns a single middleware.Responder, which is
// what the handler methods generated from the swagger definition return.
func returnsResponder(function *ast.FuncDecl) bool {
	if function.Type.Results == nil || len(function.Type.Results.List) != 1 {
		return false
	}
	selector, ok := function.Type.Results.List[0].Type.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "Responder"
}

// TestSwaggerPermissionAnnotations keeps the "Permissions required:" sentence of an operation in
// api/v2.0/swagger.yaml equal to the permission its handler enforces, so that the documentation
// an API consumer reads cannot drift away from the check the code performs. Operations whose
// handler does not check one fixed permission carry no sentence, see goharbor/harbor#21734.
func TestSwaggerPermissionAnnotations(t *testing.T) {
	operations := swaggerOperations(t)
	wanted := map[string]bool{}
	for operationID := range operations {
		wanted[operationKey(operationID)] = true
	}
	methods := handlerMethods(t, wanted)

	var missing []string
	for operationID, operation := range operations {
		method, served := methods[operationKey(operationID)]
		require.True(t, served, "no handler method of this package serves the operation %q", operationID)

		var documented string
		for _, line := range strings.Split(operation.Description, "\n") {
			if line = strings.TrimSpace(line); strings.HasPrefix(line, permissionDocPrefix) {
				documented = line
			}
		}

		switch {
		case !method.checked:
			assert.Empty(t, documented, "operation %q documents a permission but %s does not require one fixed permission of every request, remove the sentence",
				operationID, method.name)
		case documented == "":
			missing = append(missing, fmt.Sprintf("  %s: %s", operationID, method.permission.docLine()))
		default:
			assert.Equal(t, method.permission.docLine(), documented,
				"operation %q documents a permission that %s does not enforce", operationID, method.name)
		}
	}

	sort.Strings(missing)
	assert.Empty(t, missing, "the handlers of these operations enforce a permission that api/v2.0/swagger.yaml does not document, add the sentence to the description of the operation:\n%s",
		strings.Join(missing, "\n"))
}
