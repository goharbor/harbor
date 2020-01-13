package quay

import (
	"fmt"
	"strings"
)

type cred struct {
	OAuth2Token       string `json:"oauth2_token"`
	AccountName       string `json:"account_name"`
	DockerCliPassword string `json:"docker_cli_password"`
}

type orgCreate struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func buildOrgURL(endpoint, orgName string) string {
	return fmt.Sprintf("%s/api/v1/organization/%s", strings.TrimRight(endpoint, "/"), orgName)
}
