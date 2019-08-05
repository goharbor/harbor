package quayio

import "fmt"

const (
	baseURL = "https://quay.io"
)

type orgCreate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func buildOrgURL(orgName string) string {
	return fmt.Sprintf("%s/api/v1/organization/%s", baseURL, orgName)
}

func buildManifestURL(repo, ref string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", baseURL, repo, ref)
}
