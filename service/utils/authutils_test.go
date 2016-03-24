package utils

import (
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/dao"
	_ "github.com/vmware/harbor/models"
)

const username string = "Tester01"
const service string = "TesterService"
const scope string = ""

func TestMain(m *testing.M) {

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

	fmt.Printf("DB_HOST: %s, DB_USR: %s, DB_PORT: %s, DB_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	os.Setenv("MYSQL_PORT_3306_TCP_ADDR", dbHost)
	os.Setenv("MYSQL_PORT_3306_TCP_PORT", dbPort)
	os.Setenv("MYSQL_USR", dbUser)
	os.Setenv("MYSQL_PWD", dbPassword)
	os.Setenv("AUTH_MODE", "db_auth")
	dao.InitDB()
	os.Exit(m.Run())

}

func TestGenTokenForUI(t *testing.T) {

	token, err := GenTokenForUI(username, service, scope)
	if err != nil {
		t.Errorf("Error occurred in GenTokenForUI: %v", err)
	}
	log.Printf("Token: %v", token)
}
