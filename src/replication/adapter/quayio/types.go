package quayio

import "fmt"

type orgCreate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func buildOrgURL(orgName string) string {
	return fmt.Sprintf("https://quay.io/api/v1/organization/%s", orgName)
}
