package ldap

import (
	//"fmt"
	//"strings"
	"os"
	"testing"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	ldapConfig "github.com/vmware/harbor/src/ui/config"
)

func TestMain(t *testing.T) {
	os.Setenv("AUTH_MODE", "ldap_auth")
	os.Setenv("LDAP_URL", "ldap://127.0.0.1")
	os.Setenv("LDAP_BASE_DN", "dc=example,dc=com")
	os.Setenv("LDAP_SEARCH_DN", "cn=admin,dc=example,dc=com")
	os.Setenv("LDAP_SEARCH_PWD", "admin")
	os.Setenv("LDAP_UID", "uid")
	os.Setenv("LDAP_SCOPE", "3")
	os.Setenv("LDAP_FILTER", "")
	os.Setenv("LDAP_CONNECT_TIMEOUT", "1")

	db := os.Getenv("DATABASE")
	defer os.Setenv("DATABASE", db)

	os.Setenv("DATABASE", "mysql")

	dbHost := os.Getenv("DB_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable DB_HOST is not set")
	}
	dbUser := os.Getenv("DB_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable DB_USR is not set")
	}
	dbPort := os.Getenv("DB_PORT")
	if len(dbPort) == 0 {
		log.Fatalf("environment variable DB_PORT is not set")
	}
	dbPassword := os.Getenv("DB_PWD")

	log.Infof("DB_HOST: %s, DB_USR: %s, DB_PORT: %s, DB_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	os.Setenv("MYSQL_HOST", dbHost)
	os.Setenv("MYSQL_PORT", dbPort)
	os.Setenv("MYSQL_USR", dbUser)
	os.Setenv("MYSQL_PWD", dbPassword)

	err := ldapConfig.Reload()
	if err != nil {
		panic(err)
	}

	dao.InitDatabase()
}

func TestAuthenticate(t *testing.T) {
	var person models.AuthModel
	var auth *Auth
	person.Principal = "test"
	person.Password = "123456"
	user, err := auth.Authenticate(person)
	if err != nil {
		t.Errorf("unexpected ldap authenticate fail: %v", err)
	}
	if user.Username != "test" {
		t.Errorf("unexpected ldap user authenticate fail: %s = %s", "user.Username", user.Username)
	}
}
