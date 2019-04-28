package client

import (
	"github.com/theupdateframework/notary/client/changelist"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/signed"
)

// Repository represents the set of options that must be supported over a TUF repo.
type Repository interface {
	// General management operations
	Initialize(rootKeyIDs []string, serverManagedRoles ...data.RoleName) error
	InitializeWithCertificate(rootKeyIDs []string, rootCerts []data.PublicKey, serverManagedRoles ...data.RoleName) error
	Publish() error

	// Target Operations
	AddTarget(target *Target, roles ...data.RoleName) error
	RemoveTarget(targetName string, roles ...data.RoleName) error
	ListTargets(roles ...data.RoleName) ([]*TargetWithRole, error)
	GetTargetByName(name string, roles ...data.RoleName) (*TargetWithRole, error)
	GetAllTargetMetadataByName(name string) ([]TargetSignedStruct, error)

	// Changelist operations
	GetChangelist() (changelist.Changelist, error)

	// Role operations
	ListRoles() ([]RoleWithSignatures, error)
	GetDelegationRoles() ([]data.Role, error)
	AddDelegation(name data.RoleName, delegationKeys []data.PublicKey, paths []string) error
	AddDelegationRoleAndKeys(name data.RoleName, delegationKeys []data.PublicKey) error
	AddDelegationPaths(name data.RoleName, paths []string) error
	RemoveDelegationKeysAndPaths(name data.RoleName, keyIDs, paths []string) error
	RemoveDelegationRole(name data.RoleName) error
	RemoveDelegationPaths(name data.RoleName, paths []string) error
	RemoveDelegationKeys(name data.RoleName, keyIDs []string) error
	ClearDelegationPaths(name data.RoleName) error

	// Witness and other re-signing operations
	Witness(roles ...data.RoleName) ([]data.RoleName, error)

	// Key Operations
	RotateKey(role data.RoleName, serverManagesKey bool, keyList []string) error

	GetCryptoService() signed.CryptoService
	SetLegacyVersions(int)
	GetGUN() data.GUN
}
