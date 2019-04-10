package oidc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSecretVerifyError(t *testing.T) {
	sve := &SecretVerifyError{cause: fmt.Errorf("myerror")}
	assert.Equal(t, "failed to verify the secret: myerror", sve.Error())
	err := verifyError(fmt.Errorf("myerror"))
	assert.Equal(t, sve, err)
}

func TestDefaultManagerGetEncryptKey(t *testing.T) {
	d := &defaultManager{}
	k, err := d.getEncryptKey()
	assert.Nil(t, err)
	assert.Equal(t, "naa4JtarA1Zsc3uY", k)
	d2 := &defaultManager{key: "oldkey"}
	k2, err := d2.getEncryptKey()
	assert.Nil(t, err)
	assert.Equal(t, "oldkey", k2)
}

func TestPkgVerifySecret(t *testing.T) {
	SetHardcodeVerifierForTest("secret")
	assert.Nil(t, VerifySecret(context.Background(), 1, "secret"))
	assert.NotNil(t, VerifySecret(context.Background(), 1, "not-the-secret"))
}
