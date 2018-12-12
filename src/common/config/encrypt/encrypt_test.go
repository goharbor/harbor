package encrypt

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	secret := []byte("9TXCcHgNAAp1aSHh")
	filename, err := ioutil.TempFile(os.TempDir(), "keyfile")
	err = ioutil.WriteFile(filename.Name(), secret, 0644)
	if err != nil {
		fmt.Printf("failed to create temp key file\n")
	}

	defer os.Remove(filename.Name())

	os.Setenv("KEY_PATH", filename.Name())

	ret := m.Run()
	os.Exit(ret)
}

func TestEncryptDecrypt(t *testing.T) {
	password := "zhu888jie"
	encrypted, err := Instance().Encrypt(password)
	if err != nil {
		t.Errorf("Failed to decrypt password, error %v", err)
	}
	decrypted, err := Instance().Decrypt(encrypted)
	if err != nil {
		t.Errorf("Failed to decrypt password, error %v", err)
	}
	assert.NotEqual(t, password, encrypted)
	assert.Equal(t, password, decrypted)
}
