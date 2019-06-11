package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/testing/apitests/api-testing/client"
	"github.com/goharbor/harbor/src/testing/apitests/api-testing/models"
)

// UserUtil : For user related
type UserUtil struct {
	rootURI       string
	testingClient *client.APIClient
}

// NewUserUtil : Constructor
func NewUserUtil(rootURI string, httpClient *client.APIClient) *UserUtil {
	if len(strings.TrimSpace(rootURI)) == 0 || httpClient == nil {
		return nil
	}

	return &UserUtil{
		rootURI:       rootURI,
		testingClient: httpClient,
	}
}

// CreateUser : Create user
func (uu *UserUtil) CreateUser(username, password string) error {
	if len(strings.TrimSpace(username)) == 0 ||
		len(strings.TrimSpace(password)) == 0 {
		return errors.New("Username and password required for creating user")
	}

	u := models.User{
		Username: username,
		Password: password,
		Email:    username + "@vmware.com",
		RealName: username + "pks",
		Comment:  "testing",
	}

	body, err := json.Marshal(&u)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s", uu.rootURI, "/api/users")
	if err := uu.testingClient.Post(url, body); err != nil {
		return err
	}

	return nil
}

// DeleteUser : Delete testing account
func (uu *UserUtil) DeleteUser(username string) error {
	uid := uu.GetUserID(username)
	if uid == -1 {
		return fmt.Errorf("Failed to get user with name %s", username)
	}

	url := fmt.Sprintf("%s%s%d", uu.rootURI, "/api/users/", uid)
	if err := uu.testingClient.Delete(url); err != nil {
		return err
	}

	return nil
}

// GetUsers : Get users
// If name specified, then return that one
func (uu *UserUtil) GetUsers(name string) ([]models.ExistingUser, error) {
	url := fmt.Sprintf("%s%s", uu.rootURI, "/api/users")
	if len(strings.TrimSpace(name)) > 0 {
		url = url + "?username=" + name
	}

	data, err := uu.testingClient.Get(url)
	if err != nil {
		return nil, err
	}

	var users []models.ExistingUser
	if err = json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserID : Get user ID
// If user with the username is not existing, then return -1
func (uu *UserUtil) GetUserID(username string) int {
	if len(strings.TrimSpace(username)) == 0 {
		return -1
	}

	users, err := uu.GetUsers(username)
	if err != nil {
		return -1
	}

	if len(users) == 0 {
		return -1
	}

	for _, u := range users {
		if u.Username == username {
			return u.ID
		}
	}

	return -1
}
