// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkcs11

// These tests depend on SoftHSM and the library being in
// in /usr/lib/softhsm/libsofthsm.so

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"
)

/*
This test supports the following environment variables:

* SOFTHSM_LIB: complete path to libsofthsm.so
* SOFTHSM_TOKENLABEL
* SOFTHSM_PRIVKEYLABEL
* SOFTHSM_PIN
*/

func setenv(t *testing.T) *Ctx {
	lib := "/usr/lib/softhsm/libsofthsm.so"
	if x := os.Getenv("SOFTHSM_LIB"); x != "" {
		lib = x
	}
	t.Logf("loading %s", lib)
	p := New(lib)
	if p == nil {
		t.Fatal("Failed to init lib")
	}
	return p
}

func TestSetenv(t *testing.T) {
	wd, _ := os.Getwd()
	os.Setenv("SOFTHSM_CONF", wd+"/softhsm.conf")

	lib := "/usr/lib/softhsm/libsofthsm.so"
	if x := os.Getenv("SOFTHSM_LIB"); x != "" {
		lib = x
	}
	p := New(lib)
	if p == nil {
		t.Fatal("Failed to init pkcs11")
	}
	p.Destroy()
	return
}

func getSession(p *Ctx, t *testing.T) SessionHandle {
	if e := p.Initialize(); e != nil {
		t.Fatalf("init error %s\n", e)
	}
	slots, e := p.GetSlotList(true)
	if e != nil {
		t.Fatalf("slots %s\n", e)
	}
	session, e := p.OpenSession(slots[0], CKF_SERIAL_SESSION|CKF_RW_SESSION)
	if e != nil {
		t.Fatalf("session %s\n", e)
	}
	if e := p.Login(session, CKU_USER, pin); e != nil {
		t.Fatalf("user pin %s\n", e)
	}
	return session
}

func TestInitialize(t *testing.T) {
	p := setenv(t)
	if e := p.Initialize(); e != nil {
		t.Fatalf("init error %s\n", e)
	}
	p.Finalize()
	p.Destroy()
}

func TestNew(t *testing.T) {
	if p := New(""); p != nil {
		t.Fatalf("init should have failed, got %s\n", p)
	}
	if p := New("/does/not/exist"); p != nil {
		t.Fatalf("init should have failed, got %s\n", p)
	}
}

func finishSession(p *Ctx, session SessionHandle) {
	p.Logout(session)
	p.CloseSession(session)
	p.Finalize()
	p.Destroy()
}

func TestGetInfo(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)
	info, err := p.GetInfo()
	if err != nil {
		t.Fatalf("non zero error %s\n", err)
	}
	if info.ManufacturerID != "SoftHSM" {
		t.Fatal("ID should be SoftHSM")
	}
	t.Logf("%+v\n", info)
}

func TestFindObject(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)

	tokenLabel:= "TestFindObject"

	// There are 2 keys in the db with this tag
	generateRSAKeyPair(t, p, session, tokenLabel, false)

	template := []*Attribute{NewAttribute(CKA_LABEL, tokenLabel)}
	if e := p.FindObjectsInit(session, template); e != nil {
		t.Fatalf("failed to init: %s\n", e)
	}
	obj, b, e := p.FindObjects(session, 2)
	if e != nil {
		t.Fatalf("failed to find: %s %v\n", e, b)
	}
	if e := p.FindObjectsFinal(session); e != nil {
		t.Fatalf("failed to finalize: %s\n", e)
	}
	if len(obj) != 2 {
		t.Fatal("should have found two objects")
	}
}

func TestGetAttributeValue(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)

	pbk, _ := generateRSAKeyPair(t, p, session, "GetAttributeValue", false)

	template := []*Attribute{
		NewAttribute(CKA_PUBLIC_EXPONENT, nil),
		NewAttribute(CKA_MODULUS_BITS, nil),
		NewAttribute(CKA_MODULUS, nil),
		NewAttribute(CKA_LABEL, nil),
	}
	attr, err := p.GetAttributeValue(session, ObjectHandle(pbk), template)
	if err != nil {
		t.Fatalf("err %s\n", err)
	}
	for i, a := range attr {
		t.Logf("attr %d, type %d, valuelen %d", i, a.Type, len(a.Value))
		if a.Type == CKA_MODULUS {
			mod := big.NewInt(0)
			mod.SetBytes(a.Value)
			t.Logf("modulus %s\n", mod.String())
		}
	}
}

func TestDigest(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)
	e := p.DigestInit(session, []*Mechanism{NewMechanism(CKM_SHA_1, nil)})
	if e != nil {
		t.Fatalf("DigestInit: %s\n", e)
	}

	hash, e := p.Digest(session, []byte("this is a string"))
	if e != nil {
		t.Fatalf("digest: %s\n", e)
	}
	hex := ""
	for _, d := range hash {
		hex += fmt.Sprintf("%x", d)
	}
	// Teststring create with: echo -n "this is a string" | sha1sum
	if hex != "517592df8fec3ad146a79a9af153db2a4d784ec5" {
		t.Fatalf("wrong digest: %s", hex)
	}
}

func TestDigestUpdate(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)
	if e := p.DigestInit(session, []*Mechanism{NewMechanism(CKM_SHA_1, nil)}); e != nil {
		t.Fatalf("DigestInit: %s\n", e)
	}
	if e := p.DigestUpdate(session, []byte("this is ")); e != nil {
		t.Fatalf("DigestUpdate: %s\n", e)
	}
	if e := p.DigestUpdate(session, []byte("a string")); e != nil {
		t.Fatalf("DigestUpdate: %s\n", e)
	}
	hash, e := p.DigestFinal(session)
	if e != nil {
		t.Fatalf("DigestFinal: %s\n", e)
	}
	hex := ""
	for _, d := range hash {
		hex += fmt.Sprintf("%x", d)
	}
	// Teststring create with: echo -n "this is a string" | sha1sum
	if hex != "517592df8fec3ad146a79a9af153db2a4d784ec5" {
		t.Fatalf("wrong digest: %s", hex)
	}
}


/*
Purpose: Generate RSA keypair with a given name and persistence.
Inputs: test object
	context
	session handle
	tokenLabel: string to set as the token labels
	tokenPersistent: boolean. Whether or not the token should be
			session based or persistent. If false, the
			token will not be saved in the HSM and is
			destroyed upon termination of the session.
Outputs: creates persistent or ephemeral tokens within the HSM.
Returns: object handles for public and private keys. Fatal on error.
*/
func generateRSAKeyPair(t *testing.T, p *Ctx, session SessionHandle, tokenLabel string, tokenPersistent bool) (ObjectHandle, ObjectHandle) {
	/*
		inputs: test object, context, session handle
			tokenLabel: string to set as the token labels
			tokenPersistent: boolean. Whether or not the token should be
					session based or persistent. If false, the
					token will not be saved in the HSM and is
					destroyed upon termination of the session.
		outputs: creates persistent or ephemeral tokens within the HSM.
		returns: object handles for public and private keys.
	*/

	publicKeyTemplate := []*Attribute{
		NewAttribute(CKA_CLASS, CKO_PUBLIC_KEY),
		NewAttribute(CKA_KEY_TYPE, CKK_RSA),
		NewAttribute(CKA_TOKEN, tokenPersistent),
		NewAttribute(CKA_VERIFY, true),
		NewAttribute(CKA_PUBLIC_EXPONENT, []byte{1, 0, 1}),
		NewAttribute(CKA_MODULUS_BITS, 2048),
		NewAttribute(CKA_LABEL, tokenLabel),
	}
	privateKeyTemplate := []*Attribute{
		NewAttribute(CKA_TOKEN, tokenPersistent),
		NewAttribute(CKA_SIGN, true),
		NewAttribute(CKA_LABEL, tokenLabel),
		NewAttribute(CKA_SENSITIVE, true),
		NewAttribute(CKA_EXTRACTABLE, true),
	}
	pbk, pvk, e := p.GenerateKeyPair(session,
		[]*Mechanism{NewMechanism(CKM_RSA_PKCS_KEY_PAIR_GEN, nil)},
		publicKeyTemplate, privateKeyTemplate)
	if e != nil {
		t.Fatalf("failed to generate keypair: %s\n", e)
	}

	return pbk, pvk
}

func TestGenerateKeyPair(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)
	tokenLabel := "TestGenerateKeyPair"
	generateRSAKeyPair(t, p, session, tokenLabel, false)
}

func TestSign(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)

	tokenLabel := "TestSign"
	_, pvk := generateRSAKeyPair(t, p, session, tokenLabel, false)

	p.SignInit(session, []*Mechanism{NewMechanism(CKM_SHA1_RSA_PKCS, nil)}, pvk)
	_, e := p.Sign(session, []byte("Sign me!"))
	if e != nil {
		t.Fatalf("failed to sign: %s\n", e)
	}
}

/* destroyObject
	Purpose: destroy and object from the HSM
	Inputs: test handle
		session handle
		searchToken: String containing the token label to search for.
		class: Key type (CKO_PRIVATE_KEY or CKO_PUBLIC_KEY) to remove.
	Outputs: removes object from HSM
	Returns: Fatal error on failure.
*/
func destroyObject(t *testing.T, p *Ctx, session SessionHandle, searchToken string, class uint) (err error){
	template := []*Attribute{
		NewAttribute(CKA_LABEL, searchToken),
		NewAttribute(CKA_CLASS, class)}

	if e := p.FindObjectsInit(session, template); e != nil {
		t.Fatalf("failed to init: %s\n", e)
	}
	obj, _, e := p.FindObjects(session, 1)
	if e != nil || len(obj) == 0 {
		t.Fatalf("failed to find objects")
	}
	if e := p.FindObjectsFinal(session); e != nil {
		t.Fatalf("failed to finalize: %s\n", e)
	}

	if e := p.DestroyObject(session, obj[0]); e != nil {
		t.Fatalf("DestroyObject failed: %s\n", e)
	}
	return
}

// Create and destroy persistent keys
func TestDestroyObject(t *testing.T) {
	p := setenv(t)
	session := getSession(p, t)
	defer finishSession(p, session)

	generateRSAKeyPair(t, p, session, "TestDestroyKey", true)
	if e := destroyObject(t, p, session, "TestDestroyKey", CKO_PUBLIC_KEY); e != nil {
		t.Fatalf("Failed to destroy object: %s\n", e)
	}
	if e := destroyObject(t, p, session, "TestDestroyKey", CKO_PRIVATE_KEY); e != nil {
		t.Fatalf("Failed to destroy object: %s\n", e)
	}

}

// ExampleSign shows how to sign some data with a private key.
// Note: error correction is not implemented in this example.
func ExampleSign() {
	lib := "/usr/lib/softhsm/libsofthsm.so"
	if x := os.Getenv("SOFTHSM_LIB"); x != "" {
		lib = x
	}
	p := New(lib)
	if p == nil {
		log.Fatal("Failed to init lib")
	}

	p.Initialize()
	defer p.Destroy()
	defer p.Finalize()
	slots, _ := p.GetSlotList(true)
	session, _ := p.OpenSession(slots[0], CKF_SERIAL_SESSION|CKF_RW_SESSION)
	defer p.CloseSession(session)
	p.Login(session, CKU_USER, "1234")
	defer p.Logout(session)
	publicKeyTemplate := []*Attribute{
		NewAttribute(CKA_CLASS, CKO_PUBLIC_KEY),
		NewAttribute(CKA_KEY_TYPE, CKK_RSA),
		NewAttribute(CKA_TOKEN, false),
		NewAttribute(CKA_ENCRYPT, true),
		NewAttribute(CKA_PUBLIC_EXPONENT, []byte{3}),
		NewAttribute(CKA_MODULUS_BITS, 1024),
		NewAttribute(CKA_LABEL, "ExampleSign"),
	}
	privateKeyTemplate := []*Attribute{
		NewAttribute(CKA_CLASS, CKO_PRIVATE_KEY),
		NewAttribute(CKA_KEY_TYPE, CKK_RSA),
		NewAttribute(CKA_TOKEN, false),
		NewAttribute(CKA_PRIVATE, true),
		NewAttribute(CKA_SIGN, true),
		NewAttribute(CKA_LABEL, "ExampleSign"),
	}
	_, priv, err := p.GenerateKeyPair(session,
		[]*Mechanism{NewMechanism(CKM_RSA_PKCS_KEY_PAIR_GEN, nil)},
		publicKeyTemplate, privateKeyTemplate)
	if err != nil {
		log.Fatal(err)
	}
	p.SignInit(session, []*Mechanism{NewMechanism(CKM_SHA1_RSA_PKCS, nil)}, priv)
	// Sign something with the private key.
	data := []byte("Lets sign this data")

	_, err = p.Sign(session, data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("It works!")
	// Output: It works!
}
// Copyright 2013 Miek Gieben. All rights reserved.
