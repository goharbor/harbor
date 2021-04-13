package oidc

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretVerifyError(t *testing.T) {
	sve := &SecretVerifyError{cause: fmt.Errorf("myerror")}
	assert.Equal(t, "failed to verify the secret: myerror", sve.Error())
	err := verifyError(fmt.Errorf("myerror"))
	assert.Equal(t, sve, err)
}

func TestGetEncryptKey(t *testing.T) {
	kl := &keyGetter{}
	k, err := kl.encryptKey()
	assert.Nil(t, err)
	assert.Equal(t, "naa4JtarA1Zsc3uY", k)
	kl2 := &keyGetter{key: "oldkey"}
	k2, err := kl2.encryptKey()
	assert.Nil(t, err)
	assert.Equal(t, "oldkey", k2)
}

func TestPkgVerifySecret(t *testing.T) {
	SetHardcodeVerifierForTest("secret")
	u, err := VerifySecret(context.Background(), "user", "secret")
	assert.Nil(t, err)
	assert.Equal(t, "user", u.Username)
	u2, err2 := VerifySecret(context.Background(), "user2", "not-the-secret")
	assert.NotNil(t, err2)
	assert.Nil(t, u2)
}
