package dtr

type Accounts struct {
	Accounts []Account `json:"accounts"`
}

// accounts describes a namespace/org/account in DTR
// DTR calls them namespaces/orgs however they are
// accessed under an accounts API, keep that straight.
type Account struct {
	Name       string `json:"name,omitempty"`
	ID         string `json:"id,omitempty"`
	FullName   string `json:"fullName,omitempty"`
	IsOrg      bool   `json:"isOrg,omitempty"`
	IsAdmin    bool   `json:"isAdmin,omitempty"`
	IsActive   bool   `json:"isActive,omitempty"`
	IsImported bool   `json:"isImported,omitempty"`
}

// New namespace format
type namespaceCreate struct {
	FullName   string `json:"fullName,omitempty"`
	IsOrg      bool   `json:"isOrg,omitempty"`
	IsAdmin    bool   `json:"isAdmin,omitempty"`
	IsActive   bool   `json:"isActive,omitempty"`
	Name       string `json:"name,omitempty"`
	Password   string `json:"password,omitempty"`
	SearchLDAP bool   `json:"searchLDAP,omitempty"`
}

// Defaults values for a new namespace
func newDefaultDTRNamespace(name string) *namespaceCreate {
	return &namespaceCreate{
		FullName: name,
		IsOrg:    true,
		Name:     name,
	}
}

type Repositories struct {
	Repositories []Repository `json:"repositories"`
}

// repository describes a repository in DTR
type Repository struct {
	EnableManifestLists bool   `json:"enableManifestLists,omitempty`
	ID                  string `json:"id,omitempty"`
	ImmutableTags       bool   `json:"immutableTags,omitempty`
	LongDescription     string `json:"longDescription,omitempty"`
	Name                string `json:"name"`
	Namespace           string `json:"namespace"`
	NamespaceType       string `json:"namespaceType,omitempty"`
	Pulls               int64  `json:"pulls,omitempty`
	Pushes              int64  `json:"pushes,omitempty`
	ScanOnPush          bool   `json:"scanOnPush,omitempty`
	ShortDescription    string `json:"shortDescription,omitempty"`
	TagLimit            int64  `json:"tagLimit,omitempty`
	Visibility          string `json:"visibility,omitempty"`
}

// repositoryCreate describes the format for a new repository in DTR
type repositoryCreate struct {
	EnableManifestLists bool   `json:"enableManifestLists,omitempty`
	ImmutableTags       bool   `json:"immutableTags,omitempty`
	LongDescription     string `json:"longDescription,omitempty"`
	Name                string `json:"name"`
	ScanOnPush          bool   `json:"scanOnPush,omitempty`
	ShortDescription    string `json:"shortDescription,omitempty"`
	TagLimit            int64  `json:"tagLimit,omitempty`
	Visibility          string `json:"visibility,omitempty"`
}

// Defaults values for a new repository
func newDefaultDTRRepository(name string) *repositoryCreate {
	return &repositoryCreate{
		EnableManifestLists: true,
		ImmutableTags:       false,
		Name:                name,
		ScanOnPush:          false,
		TagLimit:            0,
		Visibility:          "private",
	}
}

// tag describes a tag in DTR
type Tag struct {
	Author       string `json:"author,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	Digest       string `json:"digest,omitempty"`
	HashMismatch bool   `json:"hashMismatch,omitempty`
	InNotary     bool   `json:"inNotary,omitempty`
	Name         string `json:"name"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}
