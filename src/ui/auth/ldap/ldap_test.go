package ldap

import (
	//"fmt"
	//"strings"
	"os"
	"testing"

	"github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/test"
	uiConfig "github.com/vmware/harbor/src/ui/config"
)

var adminServerLdapTestConfig = map[string]interface{}{
	config.ExtEndpoint:   "host01.com",
	config.AUTHMode:      "ldap_auth",
	config.DatabaseType:  "mysql",
	config.MySQLHost:     "127.0.0.1",
	config.MySQLPort:     3306,
	config.MySQLUsername: "root",
	config.MySQLPassword: "root123",
	config.MySQLDatabase: "registry",
	config.SQLiteFile:    "/tmp/registry.db",
	//config.SelfRegistration: true,
	config.LDAPURL:       "ldap://127.0.0.1",
	config.LDAPSearchDN:  "cn=admin,dc=example,dc=com",
	config.LDAPSearchPwd: "admin",
	config.LDAPBaseDN:    "dc=example,dc=com",
	config.LDAPUID:       "uid",
	config.LDAPFilter:    "",
	config.LDAPScope:     3,
	config.LDAPTimeout:   30,
	//	config.TokenServiceURL:            "",
	//	config.RegistryURL:                "",
	//	config.EmailHost:                  "",
	//	config.EmailPort:                  25,
	//	config.EmailUsername:              "",
	//	config.EmailPassword:              "password",
	//	config.EmailFrom:                  "from",
	//	config.EmailSSL:                   true,
	//	config.EmailIdentity:              "",
	//	config.ProjectCreationRestriction: config.ProCrtRestrAdmOnly,
	//	config.VerifyRemoteCert:           false,
	//	config.MaxJobWorkers:              3,
	//	config.TokenExpiration:            30,
	config.CfgExpiration: 5,
	//	config.JobLogDir:                  "/var/log/jobs",
	//	config.UseCompressedJS:            true,
	config.AdminInitialPassword: "password",
}

func TestMain(t *testing.T) {
	server, err := test.NewAdminserver(adminServerLdapTestConfig)
	if err != nil {
		t.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	if err := os.Setenv("ADMIN_SERVER_URL", server.URL); err != nil {
		t.Fatalf("failed to set env %s: %v", "ADMIN_SERVER_URL", err)
	}

	secretKeyPath := "/tmp/secretkey"
	_, err = test.GenerateKey(secretKeyPath)
	if err != nil {
		t.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		t.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}

	if err := uiConfig.Init(); err != nil {
		t.Fatalf("failed to initialize configurations: %v", err)
	}

	//	if err := uiConfig.Load(); err != nil {
	//		t.Fatalf("failed to load configurations: %v", err)
	//	}

	//	mode, err := uiConfig.AuthMode()
	//	if err != nil {
	//		t.Fatalf("failed to get auth mode: %v", err)
	//	}

	database, err := uiConfig.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
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
