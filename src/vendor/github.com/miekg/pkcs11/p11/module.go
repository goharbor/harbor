// Package p11 wraps `miekg/pkcs11` to make it easier to use and more idiomatic
// to Go, as compared with the more straightforward C wrapper that
// `miekg/pkcs11` presents. All types are safe to use concurrently.
//
// To use, first you open a module (a dynamically loaded library) by providing
// its path on your filesystem. This module is typically provided by
// the maker of your HSM, smartcard, or other cryptographic hardware, or
// sometimes by your operating system. Common module filenames are
// opensc-pkcs11.so, libykcs11.so, and libsofthsm2.so (you'll have to find the
// exact location).
//
// Once you've opened a Module, you can list the slots available with that
// module. Each slot may or may not contain a token. For instance, if you have a
// smartcard reader, that's a slot; if there's a smartcard in it, that's the
// token. Using this package, you can iterate through slots and check their
// information, and the information about tokens in them.
//
// Once you've found the slot with the token you want to use, you can open a
// Session with that token using OpenSession. Almost all operations require
// a session. Sessions use a sync.Mutex to ensure only one operation is active on
// them at a given time, as required by PKCS#11. If you want to get full
// performance out of a multi-core HSM, you will need to create multiple
// sessions.
//
// Once you've got a session, you can login to it. This is not necessary if you
// only want to access non-sensitive data, like certificates and public keys.
// However, to use any secret keys on a token, you'll need to login.
//
// Many operations, like FindObjects, return Objects. These represent pieces of
// data that exist on the token, referring to them by a numeric handle. With
// objects representing private keys, you can perform operations like signing
// and decrypting; with public keys and certificates you can extract their
// values.
//
// To summarize, a typical workflow might look like:
//
//   module, err := p11.OpenModule("/path/to/module.so")
//   if err != nil {
//     return err
//   }
//   slots, err := module.Slots()
//   if err != nil {
//     return err
//   }
//   // ... find the appropriate slot, then ...
//   session, err := slots[0].OpenSession()
//   if err != nil {
//     return err
//   }
//   privateKeyObject, err := session.FindObject(...)
//   if err != nil {
//     return err
//   }
//   privateKey := p11.PrivateKey(privateKeyObject)
//   signature, err := privateKey.Sign(..., []byte{"hello"})
//   if err != nil {
//     return err
//   }
package p11

import (
	"fmt"
	"sync"

	"github.com/miekg/pkcs11"
)

var modules = make(map[string]Module)
var modulesMu sync.Mutex

// OpenModule loads a PKCS#11 module (a .so file or dynamically loaded library).
// It's an error to load a PKCS#11 module multiple times, so this package
// will return a previously loaded Module for the same path if possible.
// Note that there is no facility to unload a module ("finalize" in PKCS#11
// parlance). In general, modules will be unloaded at the end of the process.
// The only place where you are likely to need to explicitly unload a module is
// if you fork your process. If you need to fork, you may want to use the
// lower-level `pkcs11` package.
func OpenModule(path string) (Module, error) {
	modulesMu.Lock()
	defer modulesMu.Unlock()
	module, ok := modules[path]
	if ok {
		return module, nil
	}

	newCtx := pkcs11.New(path)
	if newCtx == nil {
		return Module{}, fmt.Errorf("failed to load module %q", path)
	}

	err := newCtx.Initialize()
	if err != nil {
		return Module{}, fmt.Errorf("failed to initialize module: %s", err)
	}

	modules[path] = Module{newCtx}
	return modules[path], nil
}

// Module represents a PKCS#11 module, and can be used to create Sessions.
type Module struct {
	ctx *pkcs11.Ctx
}

// Info returns general information about the module.
func (m Module) Info() (pkcs11.Info, error) {
	return m.ctx.GetInfo()
}

// Slots returns all available Slots that have a token present.
func (m Module) Slots() ([]Slot, error) {
	ids, err := m.ctx.GetSlotList(true)
	if err != nil {
		return nil, err
	}
	result := make([]Slot, len(ids))
	for i, id := range ids {
		result[i] = Slot{
			ctx: m.ctx,
			id:  id,
		}
	}
	return result, nil
}
