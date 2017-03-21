package http

import (
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"fmt"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/common/utils/log"
	"net/url"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"errors"
)

// Auth implements Authenticator interface to authenticate user against LSN.
type Auth struct{}

type JsonResponse struct {
	Success bool	`json:"success"`
	Message string	`json:"message"`
}

// Authenticate calls dao to authenticate user.
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	resp, err := http.PostForm(config.HttpAuth(),
		url.Values{"username": {m.Principal}, "password": {m.Password}})
	failOnError(err, "Failed to connect to auth api")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	failOnError(err, "Failed to read response body")
	if err != nil {
		return nil, err
	}
	jsonResponse := JsonResponse{}
	err = json.Unmarshal(body, &jsonResponse)
	failOnError(err, fmt.Sprintf("Failed to parse response json: %s", string(body)))
	if err != nil {
		return nil, err
	}
	if !jsonResponse.Success {
		err = errors.New(jsonResponse.Message)
		return nil, err
	}

	u := models.User{}
	u.Username = m.Principal
	exist, err := dao.UserExists(u, "username")
	if err != nil {
		return nil, err
	}

	if exist {
		currentUser, err := dao.GetUser(u)
		if err != nil {
			return nil, err
		}
		return currentUser, nil
	} else {
		u.Realname = m.Principal
		u.Password = "12345678HttP"
		u.Comment = "registered from HTTP."
		if u.Email == "" {
			u.Email = u.Username
		}
		userID, err := dao.Register(u)
		if err != nil {
			return nil, err
		}
		u.UserID = int(userID)
	}
	return &u, nil
}

const authMode string = "http_auth"

func init() {
	auth.Register(authMode, &Auth{})
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
