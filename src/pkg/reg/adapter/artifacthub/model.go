package artifacthub

// PackageResponse ...
type PackageResponse struct {
	Data     PackageData `json:"data"`
	Metadata Metadata    `json:"metadata"`
}

// PackageData ...
type PackageData struct {
	Packages []*Package `json:"packages"`
}

// Package ...
type Package struct {
	PackageID      string      `json:"package_id"`
	Name           string      `json:"name"`
	NormalizedName string      `json:"normalized_name"`
	Repository     *Repository `json:"repository"`
}

// Repository ...
type Repository struct {
	Kind         int    `json:"kind"`
	Name         string `json:"name"`
	RepositoryID string `json:"repository_id"`
}

// PackageDetail ...
type PackageDetail struct {
	PackageID         string           `json:"package_id"`
	Name              string           `json:"name"`
	NormalizedName    string           `json:"normalized_name"`
	Version           string           `json:"version"`
	AppVersion        string           `json:"app_version"`
	Repository        RepositoryDetail `json:"repository"`
	AvailableVersions []*Version       `json:"available_versions,omitempty"`
}

// RepositoryDetail ...
type RepositoryDetail struct {
	URL               string `json:"url"`
	Kind              int    `json:"kind"`
	Name              string `json:"name"`
	RepositoryID      string `json:"repository_id"`
	VerifiedPublisher bool   `json:"verified_publisher"`
	Official          bool   `json:"official"`
	Private           bool   `json:"private"`
}

// Version ...
type Version struct {
	Version string `json:"version"`
}

// ChartVersion ...
type ChartVersion struct {
	PackageID  string `json:"package_id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	ContentURL string `json:"content_url"`
}

// Metadata ...
type Metadata struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// Message ...
type Message struct {
	Message string `json:"message"`
}

// ChartInfo ...
type ChartInfo struct {
	Repository string `json:"repository"`
	Package    string `json:"package"`
	Version    string `json:"version"`
	ContentURL string `json:"url"`
}
