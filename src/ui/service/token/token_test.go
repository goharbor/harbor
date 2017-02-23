package token

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/stretchr/testify/assert"

	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestGetResourceActions(t *testing.T) {
	s := []string{"registry:catalog:*", "repository:10.117.4.142/notary-test/hello-world-2:pull,push"}
	expectedRA := [2]token.ResourceActions{
		token.ResourceActions{
			Type:    "registry",
			Name:    "catalog",
			Actions: []string{"*"},
		},
		token.ResourceActions{
			Type:    "repository",
			Name:    "10.117.4.142/notary-test/hello-world-2",
			Actions: []string{"pull", "push"},
		},
	}
	ra := GetResourceActions(s)
	assert.Equal(t, *ra[0], expectedRA[0], "The Resource Action mismatch")
	assert.Equal(t, *ra[1], expectedRA[1], "The Resource Action mismatch")
}

func getKeyAndCertPath() (string, string) {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Join(path.Dir(f), "test/private_key.pem"), path.Join(path.Dir(f), "test/root.crt")
}

func getPublicKey(crtPath string) (*rsa.PublicKey, error) {
	crt, err := ioutil.ReadFile(crtPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(crt)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert.PublicKey.(*rsa.PublicKey), nil
}

type harborClaims struct {
	jwt.StandardClaims
	// Private claims
	Access []*token.ResourceActions `json:"access"`
}

func TestMakeToken(t *testing.T) {
	pk, crt := getKeyAndCertPath()
	//overwrite the config values for testing.
	privateKey = pk
	expiration = 10
	ra := []*token.ResourceActions{&token.ResourceActions{
		Type:    "repository",
		Name:    "10.117.4.142/notary-test/hello-world-2",
		Actions: []string{"pull", "push"},
	}}
	svc := "harbor-registry"
	u := "tester"
	tokenString, _, _, err := MakeToken(u, svc, ra)
	if err != nil {
		t.Errorf("Error while making token: %v", err)
	}
	//t.Logf("privatekey: %s, crt: %s", tokenString, crt)
	pubKey, err := getPublicKey(crt)
	if err != nil {
		t.Errorf("Error while getting public key from cert: %s", crt)
	}
	tok, err := jwt.ParseWithClaims(tokenString, &harborClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	t.Logf("validity: %v", tok.Valid)
	if err != nil {
		t.Errorf("Error while parsing the token: %v", err)
	}
	claims := tok.Claims.(*harborClaims)
	t.Logf("claims: %+v", *claims)
	assert.Equal(t, *(claims.Access[0]), *(ra[0]), "Access mismatch")
	assert.Equal(t, claims.Audience, svc, "Audience mismatch")

}
