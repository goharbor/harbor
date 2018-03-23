// +build pkcs11

package client

import "github.com/docker/notary/trustmanager/yubikey"

// clear out all keys
func init() {
	yubikey.SetYubikeyKeyMode(0)
	if !yubikey.IsAccessible() {
		return
	}
	store, err := yubikey.NewYubiStore(nil, nil)
	if err == nil {
		for k := range store.ListKeys() {
			store.RemoveKey(k)
		}
	}
}
