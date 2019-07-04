package keystone

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/config"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
)

// Auth is the implementation of AuthenticateHelper to access keystone for authentication.
type Auth struct {
	sync.Mutex
	client     *gophercloud.ServiceClient
	domainName string
	auth.DefaultAuthenticateHelper
}

// Authenticate authenticate the user access keystone service.
func (k *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	if err := k.ensureClient(); err != nil {
		return nil, err
	}

	log.Debugf("Keystone client is %+v, endpoint is %s", k.client, k.client.IdentityEndpoint)

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: k.client.IdentityEndpoint,
		DomainName:       k.domainName,
		Username:         m.Principal,
		Password:         m.Password,
	}

	result := tokens.Create(k.client, &opts)
	if result.Err != nil {
		return nil, result.Err
	}

	token, err := result.Extract()
	if err != nil {
		return nil, err
	}

	u, err := result.ExtractUser()
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Username: u.Name,
		Comment:  "From keystone",
		Realname: u.Name,
	}

	k.client.SetToken(token.ID)
	r := users.Get(k.client, u.ID)
	ku, err := r.Extract()
	if err != nil {
		return nil, err
	}

	if v, ok := ku.Extra["email"]; ok {
		user.Email = v.(string)
	} else {
		user.Email = fmt.Sprintf("%s@keystone.placeholder", u.Name)
	}

	return user, nil
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, fill in the user model based
// on the data record of the user
func (k *Auth) OnBoardUser(u *models.User) error {
	u.Username = strings.TrimSpace(u.Username)
	if len(u.Username) == 0 {
		return fmt.Errorf("username is empty")
	}
	if len(u.Realname) == 0 {
		u.Realname = u.Username
	}
	if len(u.Email) == 0 {
		u.Email = fmt.Sprintf("%s@keystone.placeholder", u.Username)
	}
	if len(u.Comment) == 0 {
		u.Comment = "From Keystone"
	}
	return dao.OnBoardUser(u)
}

// Create a group in harbor DB, if altGroupName is not empty, take the altGroupName as groupName in harbor DB.
// func (k *Auth) OnBoardGroup(g *models.UserGroup, altGroupName string) error {
//
// }

// Get user information from keystone service
// func (k *Auth) SearchUser(username string) (*models.User, error) {
// 	if err := k.ensureClient(); err != nil {
// 		return nil, err
// 	}
// 	r := users.Get(k.client, u.ID)
// 	ku, err := r.Extract()
// 	if err != nil {
// 		return nil, err
// 	}
//
// }

// Search a group based on specific authentication
// func (k *Auth) SearchGroup(groupDN string) (*models.UserGroup, error) {
//
// }

// PostAuthenticate Update user information after authenticate, such as OnBoard or sync info etc
func (k *Auth) PostAuthenticate(u *models.User) error {
	dbUser, err := dao.GetUser(models.User{Username: u.Username})
	if err != nil {
		return err
	}
	if dbUser == nil {
		return k.OnBoardUser(u)
	}
	u.UserID = dbUser.UserID
	u.HasAdminRole = dbUser.HasAdminRole
	if len(u.Email) == 0 {
		u.Email = fmt.Sprintf("%s@keystone.placeholder", u.Username)
	}
	if len(u.Comment) == 0 {
		u.Comment = "From Keystone"
	}
	if err := dao.ChangeUserProfile(*u, "Email", "Realname"); err != nil {
		log.Warningf("Failed to update user profile, user: %s, error: %v", u.Username, err)
	}

	return nil
}

func (k *Auth) ensureClient() error {
	var cfg *Config
	settings, err := config.KeystoneSetting()
	if err != nil {
		log.Warningf("Failed to get Keystone setting, error: %v", err)
	} else {
		cfg = &Config{
			Endpoint:   settings.Endpoint,
			DomainName: settings.DomainName,
		}
	}
	log.Debugf("Keystone settings: %+v", settings)

	if k.client != nil && cfg != nil {
		return k.updateClient(cfg)
	}

	if k.client == nil && cfg != nil {
		c, err := openstack.NewClient(cfg.Endpoint)
		if err != nil {
			return err
		}
		v3Client, err := openstack.NewIdentityV3(c, gophercloud.EndpointOpts{})
		if err != nil {
			return err
		}
		k.client = v3Client
		k.domainName = cfg.DomainName
	}
	return nil
}

func (k *Auth) updateClient(cfg *Config) error {
	k.client.IdentityEndpoint = cfg.Endpoint
	k.domainName = cfg.DomainName
	return nil
}

// Config keystone configurations
type Config struct {
	Endpoint   string
	DomainName string
}

func init() {
	auth.Register(common.KeystoneAuth, &Auth{})
}
