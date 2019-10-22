package dtr

// Accounts describes the DTR accounts API response
type Accounts struct {
	Accounts []Account `json:"accounts"`
}

// Account describes the DTR account API response
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

// namespaceCreate describes the format for a new namespace in DTR
type namespaceCreate struct {
	FullName   string `json:"fullName"`
	IsOrg      bool   `json:"isOrg"`
	IsAdmin    bool   `json:"isAdmin,omitempty"`
	IsActive   bool   `json:"isActive,omitempty"`
	Name       string `json:"name"`
	Password   string `json:"password,omitempty"`
	SearchLDAP bool   `json:"searchLDAP,omitempty"`
}

// newDefaultDTRNamespace is the defaults values for a new namespace
func newDefaultDTRNamespace(name string) *namespaceCreate {
	return &namespaceCreate{
		FullName: name,
		IsOrg:    true,
		Name:     name,
	}
}

// Repositories describes the DTR repositories API response
type Repositories struct {
	Repositories []Repository `json:"repositories"`
}

// Repository describes a repository in DTR
type Repository struct {
	EnableManifestLists bool   `json:"enableManifestLists,omitempty"`
	ID                  string `json:"id,omitempty"`
	ImmutableTags       bool   `json:"immutableTags,omitempty"`
	LongDescription     string `json:"longDescription,omitempty"`
	Name                string `json:"name"`
	Namespace           string `json:"namespace"`
	NamespaceType       string `json:"namespaceType,omitempty"`
	Pulls               int64  `json:"pulls,omitempty"`
	Pushes              int64  `json:"pushes,omitempty"`
	ScanOnPush          bool   `json:"scanOnPush,omitempty"`
	ShortDescription    string `json:"shortDescription,omitempty"`
	TagLimit            int64  `json:"tagLimit,omitempty"`
	Visibility          string `json:"visibility,omitempty"`
}

// repositoryCreate describes the format for a new repository in DTR
type repositoryCreate struct {
	EnableManifestLists bool   `json:"enableManifestLists,omitempty"`
	ImmutableTags       bool   `json:"immutableTags,omitempty"`
	LongDescription     string `json:"longDescription,omitempty"`
	Name                string `json:"name"`
	ScanOnPush          bool   `json:"scanOnPush,omitempty"`
	ShortDescription    string `json:"shortDescription,omitempty"`
	TagLimit            int64  `json:"tagLimit,omitempty"`
	Visibility          string `json:"visibility"`
}

// newDefaultDTRRepository is the defaults values for a new repository
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

// Tag describes the DTR tag API response
type Tag struct {
	Author       string `json:"author,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	Digest       string `json:"digest,omitempty"`
	HashMismatch bool   `json:"hashMismatch"`
	InNotary     bool   `json:"inNotary,omitempty"`
	Name         string `json:"name"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
}
