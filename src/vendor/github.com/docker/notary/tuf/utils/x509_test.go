package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/notary"
	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

func TestCertsToKeys(t *testing.T) {
	// Get root certificate
	rootCA, err := LoadCertFromFile("../../fixtures/root-ca.crt")
	require.NoError(t, err)

	// Get intermediate certificate
	intermediateCA, err := LoadCertFromFile("../../fixtures/intermediate-ca.crt")
	require.NoError(t, err)

	// Get leaf certificate
	leafCert, err := LoadCertFromFile("../../fixtures/secure.example.com.crt")
	require.NoError(t, err)

	// Get our certList with Leaf Cert and Intermediate
	certMap := map[string]*x509.Certificate{
		"a": leafCert,
		"b": intermediateCA,
		"c": rootCA,
	}
	certList := []*x509.Certificate{
		leafCert,
		intermediateCA,
		rootCA,
	}

	// Call CertsToKeys
	keys := CertsToKeys(certMap, make(map[string][]*x509.Certificate))
	require.NotNil(t, keys)
	require.Len(t, keys, 3)

	// Call GetLeafCerts
	newKeys := GetLeafCerts(certList)
	require.NotNil(t, newKeys)
	require.Len(t, newKeys, 1)

	// Call GetIntermediateCerts (checks for certs with IsCA true)
	newKeys = GetIntermediateCerts(certList)
	require.NotNil(t, newKeys)
	require.Len(t, newKeys, 2)

	// Try calling CertToKeys on a junk leaf cert that won't fingerprint
	emptyCert := x509.Certificate{}
	// Also try changing the pre-existing leaf cert into an invalid algorithm
	leafCert.PublicKeyAlgorithm = x509.DSA
	keys = CertsToKeys(map[string]*x509.Certificate{"d": &emptyCert, "e": leafCert}, make(map[string][]*x509.Certificate))
	require.Empty(t, keys)
}

func TestNewCertificate(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.AddDate(10, 0, 0)
	cert, err := NewCertificate("docker.com/alpine", startTime, endTime)
	require.NoError(t, err)
	require.Equal(t, cert.Subject.CommonName, "docker.com/alpine")
	require.Equal(t, cert.NotBefore, startTime)
	require.Equal(t, cert.NotAfter, endTime)
}

func TestKeyOperations(t *testing.T) {
	// Generate our ED25519 private key
	edKey, err := GenerateED25519Key(rand.Reader)
	require.NoError(t, err)

	// Generate our EC private key
	ecKey, err := GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	// parse a RSA key
	block, _ := pem.Decode([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDJ8BO2/HOHLJgrb3srafbNRUD8r0SGNJFi5h7t4vxZ4F5oBW/4
O2/aZmdToinyuCm0eGguK77HAsTfSHqDUoEfuInNg7pPk4F6xa4feQzEeG6P0YaL
+VbApUdCHLBE0tVZg1SCW97+27wqIM4Cl1Tcsbb+aXfgMaOFGxlyga+a6wIDAQAB
AoGBAKDWLH2kGMfjBtghlLKBVWcs75PSbPuPRvTEYIIMNf3HrKmhGwtVG8ORqF5+
XHbLo7vv4tpTUUHkvLUyXxHVVq1oX+QqiRwTRm+ROF0/T6LlrWvTzvowTKtkRbsm
mqIYEbc+fBZ/7gEeW2ycCfE7HWgxNGvbUsK4LNa1ozJbrVEBAkEA8ML0mXyxq+cX
CwWvdXscN9vopLG/y+LKsvlKckoI/Hc0HjPyraq5Docwl2trZEmkvct1EcN8VvcV
vCtVsrAfwQJBANa4EBPfcIH2NKYHxt9cP00n74dVSHpwJYjBnbec5RCzn5UTbqd2
i62AkQREYhHZAryvBVE81JAFW3nqI9ZTpasCQBqEPlBRTXgzYXRTUfnMb1UvoTXS
Zd9cwRppHmvr/4Ve05yn+AhsjyksdouWxyMqgTxuFhy4vQ8O85Pf6fZeM4ECQCPp
Wv8H4thJplqSeGeJFSlBYaVf1SRtN0ndIBTCj+kwMaOMQXiOsiPNmfN9wG09v2Bx
YVFJ/D8uNjN4vo+tI8sCQFbtF+Qkj4uSFDZGMESF6MOgsGt1R1iCpvpMSr9h9V02
LPXyS3ozB7Deq26pEiCrFtHxw2Pb7RJO6GEqH7Dg4oU=
-----END RSA PRIVATE KEY-----`))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	rsaKey, err := RSAToPrivateKey(key)
	require.NoError(t, err)

	// Encode our ED private key
	edPEM, err := ConvertPrivateKeyToPKCS8(edKey, data.CanonicalRootRole, "", "")
	require.NoError(t, err)

	// Encode our EC private key
	ecPEM, err := ConvertPrivateKeyToPKCS8(ecKey, data.CanonicalRootRole, "", "")
	require.NoError(t, err)

	// Encode our RSA private key
	rsaPEM, err := ConvertPrivateKeyToPKCS8(rsaKey, data.CanonicalRootRole, "", "")
	require.NoError(t, err)

	// Check to see if ED key it is encoded
	stringEncodedEDKey := string(edPEM)
	require.True(t, strings.Contains(stringEncodedEDKey, "-----BEGIN PRIVATE KEY-----"))

	// Check to see the ED key type
	testKeyBlockType(t, edPEM, nil, "ed25519")

	// Check to see if EC key it is encoded
	stringEncodedECKey := string(ecPEM)
	require.True(t, strings.Contains(stringEncodedECKey, "-----BEGIN PRIVATE KEY-----"))

	// Check to see the EC key type
	testKeyBlockType(t, ecPEM, nil, "ecdsa")

	// Check to see if RSA key it is encoded
	stringEncodedRSAKey := string(rsaPEM)
	require.True(t, strings.Contains(stringEncodedRSAKey, "-----BEGIN PRIVATE KEY-----"))

	// Check to see the RSA key type
	testKeyBlockType(t, rsaPEM, nil, "rsa")

	// Try to decode garbage bytes
	_, err = ParsePEMPrivateKey([]byte("Knock knock; it's Bob."), "")
	require.Error(t, err)

	// Decode our ED Key
	decodedEDKey, err := ParsePEMPrivateKey(edPEM, "")
	require.NoError(t, err)
	require.Equal(t, edKey.Private(), decodedEDKey.Private())

	// Decode our EC Key
	decodedECKey, err := ParsePEMPrivateKey(ecPEM, "")
	require.NoError(t, err)
	require.Equal(t, ecKey.Private(), decodedECKey.Private())

	// Decode our RSA Key
	decodedRSAKey, err := ParsePEMPrivateKey(rsaPEM, "")
	require.NoError(t, err)
	require.Equal(t, rsaKey.Private(), decodedRSAKey.Private())

	// Encrypt our ED Key
	encryptedEDKey, err := ConvertPrivateKeyToPKCS8(edKey, data.CanonicalRootRole, "", "ponies")
	require.NoError(t, err)

	// Encrypt our EC Key
	encryptedECKey, err := ConvertPrivateKeyToPKCS8(ecKey, data.CanonicalRootRole, "", "ponies")
	require.NoError(t, err)

	// Encrypt our RSA Key
	encryptedRSAKey, err := ConvertPrivateKeyToPKCS8(rsaKey, data.CanonicalRootRole, "", "ponies")
	require.NoError(t, err)

	// Check to see if ED key it is encrypted
	stringEncryptedEDKey := string(encryptedEDKey)
	require.True(t, strings.Contains(stringEncryptedEDKey, "-----BEGIN ENCRYPTED PRIVATE KEY-----"))
	role, _, err := ExtractPrivateKeyAttributes(encryptedEDKey)
	require.NoError(t, err)
	require.EqualValues(t, "root", role)

	// Check to see if EC key it is encrypted
	stringEncryptedECKey := string(encryptedECKey)
	require.True(t, strings.Contains(stringEncryptedECKey, "-----BEGIN ENCRYPTED PRIVATE KEY-----"))
	role, _, err = ExtractPrivateKeyAttributes(encryptedECKey)
	require.NoError(t, err)
	require.EqualValues(t, "root", role)

	// Check to see if RSA key it is encrypted
	stringEncryptedRSAKey := string(encryptedRSAKey)
	require.True(t, strings.Contains(stringEncryptedRSAKey, "-----BEGIN ENCRYPTED PRIVATE KEY-----"))
	role, _, err = ExtractPrivateKeyAttributes(encryptedRSAKey)
	require.NoError(t, err)
	require.EqualValues(t, "root", role)

	// Decrypt our ED Key
	decryptedEDKey, err := ParsePEMPrivateKey(encryptedEDKey, "ponies")
	require.NoError(t, err)
	require.Equal(t, edKey.Private(), decryptedEDKey.Private())

	// Decrypt our EC Key
	decryptedECKey, err := ParsePEMPrivateKey(encryptedECKey, "ponies")
	require.NoError(t, err)
	require.Equal(t, ecKey.Private(), decryptedECKey.Private())

	// Decrypt our RSA Key
	decryptedRSAKey, err := ParsePEMPrivateKey(encryptedRSAKey, "ponies")
	require.NoError(t, err)
	require.Equal(t, rsaKey.Private(), decryptedRSAKey.Private())

	// quick test that gun headers are being added appropriately
	// Encrypt our RSA Key, one type of key should be enough since headers are treated the same
	testGunKey, err := ConvertPrivateKeyToPKCS8(rsaKey, data.CanonicalRootRole, "ilove", "ponies")
	require.NoError(t, err)

	testNoGunKey, err := ConvertPrivateKeyToPKCS8(rsaKey, data.CanonicalRootRole, "", "ponies")
	require.NoError(t, err)

	_, gun, err := ExtractPrivateKeyAttributes(testGunKey)
	require.NoError(t, err)
	require.EqualValues(t, "ilove", gun)

	_, gun, err = ExtractPrivateKeyAttributes(testNoGunKey)
	require.NoError(t, err)
	require.EqualValues(t, "", gun)
}

func testKeyBlockType(t *testing.T, b, password []byte, expectedKeyType string) {
	block, _ := pem.Decode(b)
	require.NotNil(t, block)

	privKey, err := ParsePKCS8ToTufKey(block.Bytes, password)
	require.NoError(t, err, "unable to parse to pkcs8")

	require.Equal(t, expectedKeyType, privKey.Algorithm(), "key type did not match")
}

// X509PublickeyID returns the public key ID of a RSA X509 key rather than the
// cert ID
func TestRSAX509PublickeyID(t *testing.T) {
	fileBytes, err := ioutil.ReadFile("../../fixtures/notary-server.key")
	require.NoError(t, err)

	privKey, err := ParsePEMPrivateKey(fileBytes, "")
	require.NoError(t, err)
	expectedTUFID := privKey.ID()

	cert, err := LoadCertFromFile("../../fixtures/notary-server.crt")
	require.NoError(t, err)

	rsaKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	require.NoError(t, err)

	sameWayTUFID := data.NewPublicKey(data.RSAKey, rsaKeyBytes).ID()

	actualTUFKey := CertToKey(cert)
	actualTUFID, err := X509PublicKeyID(actualTUFKey)
	require.NoError(t, err)
	require.Equal(t, sameWayTUFID, actualTUFID)
	require.Equal(t, expectedTUFID, actualTUFID)
}

// X509PublickeyID returns the public key ID of an ECDSA X509 key rather than
// the cert ID
func TestECDSAX509PublickeyID(t *testing.T) {
	startTime := time.Now()
	template, err := NewCertificate("something", startTime, startTime.AddDate(10, 0, 0))
	require.NoError(t, err)
	template.SignatureAlgorithm = x509.ECDSAWithSHA256
	template.PublicKeyAlgorithm = x509.ECDSA

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tufPrivKey, err := ECDSAToPrivateKey(privKey)
	require.NoError(t, err)

	derBytes, err := x509.CreateCertificate(
		rand.Reader, template, template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	tufKey := CertToKey(cert)
	tufID, err := X509PublicKeyID(tufKey)
	require.NoError(t, err)

	require.Equal(t, tufPrivKey.ID(), tufID)
}

func preserveEnv(name string) func() {
	if env, has := os.LookupEnv(name); has {
		os.Unsetenv(name)
		return func() {
			os.Setenv(name, env)
		}
	}

	return func() {}
}

func TestExtractPrivateKeyAttributes(t *testing.T) {
	testExtractPrivateKeyAttributes(t)
	testExtractPrivateKeyAttributesWithFIPS(t)
}

func testExtractPrivateKeyAttributes(t *testing.T) {
	defer preserveEnv(notary.FIPSEnvVar)()

	err := os.Unsetenv(notary.FIPSEnvVar)
	require.NoError(t, err)

	testPKCS1PEM1 := getPKCS1KeyWithRole(t, "unicorn", "rainbow")
	testPKCS1PEM2 := getPKCS1KeyWithRole(t, "docker", "")
	testPKCS8PEM1 := getPKCS8KeyWithRole(t, "fat", "panda")
	testPKCS8PEM2 := getPKCS8KeyWithRole(t, "dagger", "")

	// Try garbage bytes
	_, _, err = ExtractPrivateKeyAttributes([]byte("Knock knock; it's Bob."))
	require.Error(t, err)

	// PKCS#8
	role, gun, err := ExtractPrivateKeyAttributes(testPKCS8PEM1)
	require.NoError(t, err)
	require.EqualValues(t, data.RoleName("fat"), role)
	require.EqualValues(t, data.GUN("panda"), gun)

	role, gun, err = ExtractPrivateKeyAttributes(testPKCS8PEM2)
	require.NoError(t, err)
	require.EqualValues(t, data.RoleName("dagger"), role)
	require.EqualValues(t, data.GUN(""), gun)

	// PKCS#1
	role, gun, err = ExtractPrivateKeyAttributes(testPKCS1PEM1)
	require.NoError(t, err)
	require.EqualValues(t, data.RoleName("unicorn"), role)
	require.EqualValues(t, data.GUN("rainbow"), gun)

	role, gun, err = ExtractPrivateKeyAttributes(testPKCS1PEM2)
	require.NoError(t, err)
	require.EqualValues(t, data.RoleName("docker"), role)
	require.EqualValues(t, data.GUN(""), gun)
}

func testExtractPrivateKeyAttributesWithFIPS(t *testing.T) {
	defer preserveEnv(notary.FIPSEnvVar)()

	err := os.Setenv(notary.FIPSEnvVar, "1")
	require.NoError(t, err)

	testPKCS1PEM1 := getPKCS1KeyWithRole(t, "unicorn", "rainbow")
	testPKCS1PEM2 := getPKCS1KeyWithRole(t, "docker", "")

	// PKCS#1
	_, _, err = ExtractPrivateKeyAttributes(testPKCS1PEM1)
	require.Error(t, err)
	_, _, err = ExtractPrivateKeyAttributes(testPKCS1PEM2)
	require.Error(t, err)

	testPKCS8PEM1 := getPKCS8KeyWithRole(t, "fat", "panda")
	testPKCS8PEM2 := getPKCS8KeyWithRole(t, "dagger", "")

	// Try garbage bytes
	_, _, err = ExtractPrivateKeyAttributes([]byte("Knock knock; it's Bob."))
	require.Error(t, err)

	// PKCS#8
	role, gun, err := ExtractPrivateKeyAttributes(testPKCS8PEM1)
	require.NoError(t, err)
	require.EqualValues(t, data.RoleName("fat"), role)
	require.EqualValues(t, data.GUN("panda"), gun)

	role, gun, err = ExtractPrivateKeyAttributes(testPKCS8PEM2)
	require.NoError(t, err)
	require.EqualValues(t, data.RoleName("dagger"), role)
	require.EqualValues(t, data.GUN(""), gun)
}

func TestParsePEMPrivateKeyLegacy(t *testing.T) {
	keys := []string{
		`-----BEGIN ED25519 PRIVATE KEY-----
role: root

cVlz659cVO/F6/EOGjuXms1RvX/PmSJalv341Un6N7qmcFuchu5zBcRBZI4nDFwn
dXJONWxS4OL7kgBTbEeYtnFZc+ufXFTvxevxDho7l5rNUb1/z5kiWpb9+NVJ+je6
-----END ED25519 PRIVATE KEY-----
`,
		`-----BEGIN EC PRIVATE KEY-----
role: root

MHcCAQEEILV9cO28WyG1iMPUFXP++AtWDBJv5hQeusiUHL4EeUzRoAoGCCqGSM49
AwEHoUQDQgAEylKk0nv1In9WNEhTI39MHZhgC8OCBE+Qp/7Fbz65bzSkeFYRIz7x
122yOUaphJOboK52o6HDbs4QcIkJ0yl/NQ==
-----END EC PRIVATE KEY-----
`,
		`-----BEGIN RSA PRIVATE KEY-----
role: root

MIIEpAIBAAKCAQEAqDpGSq+s8LyhCI8KVscVFfDxMlJdZ8QthwG4/GvKXXNpjyrn
49Tl5Hn9KUAyNVpzYNU6+KbQcprFiIUAYljlC4MCHLKsSTYHlRVnPszahgzMnkBJ
wRhgjmvG9smpzFd/0typ1ToRwIJuLPmnT1xC15F7x3WdOx55lJ+WRG5tkhLflKRC
Uw8JlDEBbMbhOO1E3IS2OeCReLgMao8yalltsN7+4GghXmVCLc0Iewvkx6DMeNRD
FEFRQJpHzbU5DWU3I8hf6yDyJ7+gfSYnyYPX6mgh3IvJHqRNYdfMZHFe+/pEM+DL
dOxUpIMgCp57DkwubgoHaZuWt6ix8IB6VJi1dQIDAQABAoIBAGiQ0GBcEgMRPIc1
YhfOZyzcNxAwYh69sg7Y40MDPSYZNuPmp3zWOI9rxBB/9rVzI4RtBdrI1Yhm66GQ
Ck0XNEeThxyPcseO7eedBi/i5XGtQwKasz1zCZF9LI75irGZMbq/rlD7Z01hxVnv
VC/gCSw1Ids5ICI/LxNSnvSqLzE7x11n6beXoummFoncvQDoFrjM3PrFRxVpLppJ
g3IWGwvQIAhTNA6ahUItuPnZhncARTCaYsfTNbX/zjbWhXIcL6MGJ2dVVJPLwVpY
3pvvbCR0oJKIJrJcXqD+ax/xhmYNAU4/LcNI9tirGHRxx+uJ6/zW4f+tRF9zXCyU
G6m/oAECgYEA3qZ86b/YeilO9Oa+7HaClHWCYMXNhAGYviNHqRvRSq849QyWwCRk
qvogdBYq+5KfX87ASIZFWmgM51QkmloF0DRurit3YRR0PuTuIZr11M+kCB7+HWrd
DC+3CzYpY6hQU0fsNLBl/x8std4RhjXa6oApsJYjAyr8adY2Qg5q0DUCgYEAwWzz
nlwacNefFN6YTXWQjGZiHxvHBnXQP5OLuTI3nUNWh4/tJ3/ktsuWFkHFW7rlS2B6
avSbpycMCxD2mdHt+IlBfzAZAG1ik23SttFbfD4nnQN7NktjZwTmKI5tjCvs3HB6
JO4CcxJ/VNw/WtHlJu2oPViAy8cLOm2k3TB9eEECgYBFAaTFbchSVGs8TCfwceqW
yLTX+XZw623DwHt9VjnPw+8LRBOVCbKJq2xTjmtT/WWX9CR0Vek40/br25BcpnoW
xaloIeCmHgjJVXrYv4ZhptlYCwMHaw+Hr2Iz/11knc4HgcsbqXBzWd4pn+IejqKC
+6XwLRg86x3AT7wRTRad4QKBgQCMs62Pn54YQbFV5ApUJlYM25k62eDwIRlodfLo
t8/e1RIHsLmpxw3frr6x2AwxiwWqzDagwOjNMck/74oDIMODzIxZcept9iQD7Jqg
JDDxcuEsBVFGkJZxZQ3rqJelpHo7bJJddMlRXb5EQ6bOcOrJY43DejLOiS7wxLtt
rw1GQQKBgQCKOduk0ugZe/t/BSkqf3/vlEk4FjwErPMTpLnsVPN13QA5//XHahkm
bsCmy8401/anZGa1s6m58UswCWNhJlCtfozN/rtkgjWZGOtlc3at0MYDdObBynVg
PBV11bfmoHzDVeeuz1ztFUb3WjR7xlQe09izY3o3N6yZlTFIsqawIg==
-----END RSA PRIVATE KEY-----
`,
	}

	for _, key := range keys {
		testParsePEMPrivateKeyLegacy(t, []byte(key))
		testParsePEMPrivateKeyLegacyWithFIPS(t, []byte(key))
	}
}

func testParsePEMPrivateKeyLegacy(t *testing.T, raw []byte) {
	defer preserveEnv(notary.FIPSEnvVar)()

	err := os.Unsetenv(notary.FIPSEnvVar)
	require.NoError(t, err)

	key, err := ParsePEMPrivateKey(raw, "")
	require.NoError(t, err)
	require.NotNil(t, key.Public())
	require.NotNil(t, key.Private())
}

func testParsePEMPrivateKeyLegacyWithFIPS(t *testing.T, raw []byte) {
	defer preserveEnv(notary.FIPSEnvVar)()

	err := os.Setenv(notary.FIPSEnvVar, "1")
	require.NoError(t, err)

	// No legacy key must be accepted in FIPS mode
	_, err = ParsePEMPrivateKey(raw, "")
	require.Error(t, err)
}

func getPKCS1KeyWithRole(t *testing.T, role data.RoleName, gun data.GUN) []byte {
	var testPEM []byte
	if gun == "" {
		testPEM = []byte(fmt.Sprintf(`-----BEGIN RSA PRIVATE KEY-----
role: %s
MIIEogIBAAKCAQEAyUIXjsrWRrvPa4Bzp3VJ6uOUGPay2fUpSV8XzNxZxIG/Opdr
+k3EQi1im6WOqF3Y5AS1UjYRxNuRN+cAZeo3uS1pOTuoSupBXuchVw8s4hZJ5vXn
TRmGb+xY7tZ1ZVgPfAZDib9sRSUsL/gC+aSyprAjG/YBdbF06qKbfOfsoCEYW1OQ
82JqHzQH514RFYPTnEGpvfxWaqmFQLmv0uMxV/cAYvqtrGkXuP0+a8PknlD2obw5
0rHE56Su1c3Q42S7L51K38tpbgWOSRcTfDUWEj5v9wokkNQvyKBwbS996s4EJaZd
7r6M0h1pHnuRxcSaZLYRwgOe1VNGg2VfWzgd5QIDAQABAoIBAF9LGwpygmj1jm3R
YXGd+ITugvYbAW5wRb9G9mb6wspnwNsGTYsz/UR0ZudZyaVw4jx8+jnV/i3e5PC6
QRcAgqf8l4EQ/UuThaZg/AlT1yWp9g4UyxNXja87EpTsGKQGwTYxZRM4/xPyWOzR
mt8Hm8uPROB9aA2JG9npaoQG8KSUj25G2Qot3ukw/IOtqwN/Sx1EqF0EfCH1K4KU
a5TrqlYDFmHbqT1zTRec/BTtVXNsg8xmF94U1HpWf3Lpg0BPYT7JiN2DPoLelRDy
a/A+a3ZMRNISL5wbq/jyALLOOyOkIqa+KEOeW3USuePd6RhDMzMm/0ocp5FCwYfo
k4DDeaECgYEA0eSMD1dPGo+u8UTD8i7ZsZCS5lmXLNuuAg5f5B/FGghD8ymPROIb
dnJL5QSbUpmBsYJ+nnO8RiLrICGBe7BehOitCKi/iiZKJO6edrfNKzhf4XlU0HFl
jAOMa975pHjeCoZ1cXJOEO9oW4SWTCyBDBSqH3/ZMgIOiIEk896lSmkCgYEA9Xf5
Jqv3HtQVvjugV/axAh9aI8LMjlfFr9SK7iXpY53UdcylOSWKrrDok3UnrSEykjm7
UL3eCU5jwtkVnEXesNn6DdYo3r43E6iAiph7IBkB5dh0yv3vhIXPgYqyTnpdz4pg
3yPGBHMPnJUBThg1qM7k6a2BKHWySxEgC1DTMB0CgYAGvdmF0J8Y0k6jLzs/9yNE
4cjmHzCM3016gW2xDRgumt9b2xTf+Ic7SbaIV5qJj6arxe49NqhwdESrFohrKaIP
kM2l/o2QaWRuRT/Pvl2Xqsrhmh0QSOQjGCYVfOb10nAHVIRHLY22W4o1jk+piLBo
a+1+74NRaOGAnu1J6/fRKQKBgAF180+dmlzemjqFlFCxsR/4G8s2r4zxTMXdF+6O
3zKuj8MbsqgCZy7e8qNeARxwpCJmoYy7dITNqJ5SOGSzrb2Trn9ClP+uVhmR2SH6
AlGQlIhPn3JNzI0XVsLIloMNC13ezvDE/7qrDJ677EQQtNEKWiZh1/DrsmHr+irX
EkqpAoGAJWe8PC0XK2RE9VkbSPg9Ehr939mOLWiHGYTVWPttUcum/rTKu73/X/mj
WxnPWGtzM1pHWypSokW90SP4/xedMxludvBvmz+CTYkNJcBGCrJumy11qJhii9xp
EMl3eFOJXjIch/wIesRSN+2dGOsl7neercjMh1i9RvpCwHDx/E0=
-----END RSA PRIVATE KEY-----
`, role))
	} else {
		testPEM = []byte(fmt.Sprintf(`-----BEGIN RSA PRIVATE KEY-----
gun: %s
role: %s
MIIEogIBAAKCAQEAyUIXjsrWRrvPa4Bzp3VJ6uOUGPay2fUpSV8XzNxZxIG/Opdr
+k3EQi1im6WOqF3Y5AS1UjYRxNuRN+cAZeo3uS1pOTuoSupBXuchVw8s4hZJ5vXn
TRmGb+xY7tZ1ZVgPfAZDib9sRSUsL/gC+aSyprAjG/YBdbF06qKbfOfsoCEYW1OQ
82JqHzQH514RFYPTnEGpvfxWaqmFQLmv0uMxV/cAYvqtrGkXuP0+a8PknlD2obw5
0rHE56Su1c3Q42S7L51K38tpbgWOSRcTfDUWEj5v9wokkNQvyKBwbS996s4EJaZd
7r6M0h1pHnuRxcSaZLYRwgOe1VNGg2VfWzgd5QIDAQABAoIBAF9LGwpygmj1jm3R
YXGd+ITugvYbAW5wRb9G9mb6wspnwNsGTYsz/UR0ZudZyaVw4jx8+jnV/i3e5PC6
QRcAgqf8l4EQ/UuThaZg/AlT1yWp9g4UyxNXja87EpTsGKQGwTYxZRM4/xPyWOzR
mt8Hm8uPROB9aA2JG9npaoQG8KSUj25G2Qot3ukw/IOtqwN/Sx1EqF0EfCH1K4KU
a5TrqlYDFmHbqT1zTRec/BTtVXNsg8xmF94U1HpWf3Lpg0BPYT7JiN2DPoLelRDy
a/A+a3ZMRNISL5wbq/jyALLOOyOkIqa+KEOeW3USuePd6RhDMzMm/0ocp5FCwYfo
k4DDeaECgYEA0eSMD1dPGo+u8UTD8i7ZsZCS5lmXLNuuAg5f5B/FGghD8ymPROIb
dnJL5QSbUpmBsYJ+nnO8RiLrICGBe7BehOitCKi/iiZKJO6edrfNKzhf4XlU0HFl
jAOMa975pHjeCoZ1cXJOEO9oW4SWTCyBDBSqH3/ZMgIOiIEk896lSmkCgYEA9Xf5
Jqv3HtQVvjugV/axAh9aI8LMjlfFr9SK7iXpY53UdcylOSWKrrDok3UnrSEykjm7
UL3eCU5jwtkVnEXesNn6DdYo3r43E6iAiph7IBkB5dh0yv3vhIXPgYqyTnpdz4pg
3yPGBHMPnJUBThg1qM7k6a2BKHWySxEgC1DTMB0CgYAGvdmF0J8Y0k6jLzs/9yNE
4cjmHzCM3016gW2xDRgumt9b2xTf+Ic7SbaIV5qJj6arxe49NqhwdESrFohrKaIP
kM2l/o2QaWRuRT/Pvl2Xqsrhmh0QSOQjGCYVfOb10nAHVIRHLY22W4o1jk+piLBo
a+1+74NRaOGAnu1J6/fRKQKBgAF180+dmlzemjqFlFCxsR/4G8s2r4zxTMXdF+6O
3zKuj8MbsqgCZy7e8qNeARxwpCJmoYy7dITNqJ5SOGSzrb2Trn9ClP+uVhmR2SH6
AlGQlIhPn3JNzI0XVsLIloMNC13ezvDE/7qrDJ677EQQtNEKWiZh1/DrsmHr+irX
EkqpAoGAJWe8PC0XK2RE9VkbSPg9Ehr939mOLWiHGYTVWPttUcum/rTKu73/X/mj
WxnPWGtzM1pHWypSokW90SP4/xedMxludvBvmz+CTYkNJcBGCrJumy11qJhii9xp
EMl3eFOJXjIch/wIesRSN+2dGOsl7neercjMh1i9RvpCwHDx/E0=
-----END RSA PRIVATE KEY-----
`, gun, role))
	}

	return testPEM
}

func getPKCS8KeyWithRole(t *testing.T, role data.RoleName, gun data.GUN) []byte {
	var testPEM []byte
	if gun == "" {
		testPEM = []byte(fmt.Sprintf(`-----BEGIN PRIVATE KEY-----
role: %s
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCo23n5TrVazr9C
DniRin4uSxx9w4tDtp5WbvACE0iWw2T0l6GHPnKa6aBFmJ3GxRQtwveM+cQVub3A
KjIS7OdJpeQA5fOrpnC5dgv5l/DbmZ5SHrzAm9JgYmrw7Uj1dyeQ6jpPra4ChF7P
YOIcoeTp4eWOSJzztZQyftnBLVRgTIYXXX1kVZVfWVFsT2FQk7ei9Gw/UnslyVQd
HZxYa98SAmsoQ5YZb3I11Tk1LHsCS6Py9p6tL3vdyW9rJioqNu2RhO/WwhqiSttd
/xTanwJRlD5IWE32CU3II4UtaZZrYDeimRekVV7zqwgTBAzNsmatZhD1o6E9LZ9D
JFhKwwDXAgMBAAECggEAbqa0PV0IlqMYze6xr53zpd5uozM61XqcM8Oq35FHZhRQ
2b9riDax3zXtYu3pplGLMZmrouQhTKNU5tI/0gsQXUCqMrR9gyQkhkQHAN5CZYU7
LFEcG5OAvsx/i7XSs5gLg3kaERCdEOUxQ/AW+/BTE7iGN0D6KPH6VUSu6VoNCrTK
PmYvgta7hwebnvo65/OAc4inp9C19FUkhcNbaCKduWBgUt348+IzVEw9H8+PrdVZ
dYGfVXAsDFY3zz0ThUbaZ52XS1pCCQ1Df9bQnTgqJNc+u1xQHLYAageKS83uAbtS
nYjBFFuxeRR2FA1n8echCWQV+16Kqq31U1E2yLfWcQKBgQDSoT73pO9h/yN5myqu
XxhAC+Ndas0DTl4pmVPpybpenJerba/0KCfYpcSFHAdkXZ1DYL7U+9uh5NRul09f
WdjayFjn0vv63rwX+PGi1yPHTIv5kLvjYXJtaxzxSzQivYMPmD/7QX4tEsUkpJ8k
90vMSS/C5ieWbpFwWVvEjFbqHQKBgQDNOsTq6USE3R8p4lj9Ror/waUuWVdzZnm3
uZGJw3WzvzaXmqLSfVHntUnD8TPHgk3WZxDhrF73POMADkl9IN/JPI150/Uo6YJo
qYGoZr0mmnEZxVCkwODz5C9icnyjklcRdIRM6eljhFMQDVEacDkptsntHUyIdQZc
L2eLNUfEgwKBgHxy7UNg3lemag110rgIU8mzvHj7m3oymYw2nc/qcwVnvG17d5Tp
DPICr6R+NRfl//9JcDdjQBfdnm5hVHJgIbLS4UTH8j390GDRo+O0/dzJq4KfM4Rb
lUJ1ITqoVnuYQZG7QUJxJd330yedZLJwswZWz7N2TTmixqf9BC2TRd85AoGAN+Qh
bLhKaMSvkACMq61ifXSHP7AlGNB3pYlsEVCh5WnVvEPow9pNTAUbKbmumE7sU8+N
0WfYFQ0H5SP+74zcZTmQbfVDdvjhAw/mt64DJVg6JQKPi87bdJBYNz9mokVgYOiS
fz/Ux71pwZ1e0QxvBOU66NBp31+/c6uVT1wbR3ECgYAdye1+UPpS9Dn89g6Ks0kv
UaFKykXu7vY2uxiNqhmWzze4iq5wmIHmEwc6+rVMluXQPAME7Iya3mBmto9AHQ/n
/ka+fGoaUgAojCLZW5DZcelIETw+Dk+95vyyAUsWfAvn4nKo4/rkBXcSHlvgElzq
SorPiBWYosFB6jqUTXew2w==
-----END PRIVATE KEY-----
`, role))
	} else {
		testPEM = []byte(fmt.Sprintf(`-----BEGIN PRIVATE KEY-----
gun: %s
role: %s
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCo23n5TrVazr9C
DniRin4uSxx9w4tDtp5WbvACE0iWw2T0l6GHPnKa6aBFmJ3GxRQtwveM+cQVub3A
KjIS7OdJpeQA5fOrpnC5dgv5l/DbmZ5SHrzAm9JgYmrw7Uj1dyeQ6jpPra4ChF7P
YOIcoeTp4eWOSJzztZQyftnBLVRgTIYXXX1kVZVfWVFsT2FQk7ei9Gw/UnslyVQd
HZxYa98SAmsoQ5YZb3I11Tk1LHsCS6Py9p6tL3vdyW9rJioqNu2RhO/WwhqiSttd
/xTanwJRlD5IWE32CU3II4UtaZZrYDeimRekVV7zqwgTBAzNsmatZhD1o6E9LZ9D
JFhKwwDXAgMBAAECggEAbqa0PV0IlqMYze6xr53zpd5uozM61XqcM8Oq35FHZhRQ
2b9riDax3zXtYu3pplGLMZmrouQhTKNU5tI/0gsQXUCqMrR9gyQkhkQHAN5CZYU7
LFEcG5OAvsx/i7XSs5gLg3kaERCdEOUxQ/AW+/BTE7iGN0D6KPH6VUSu6VoNCrTK
PmYvgta7hwebnvo65/OAc4inp9C19FUkhcNbaCKduWBgUt348+IzVEw9H8+PrdVZ
dYGfVXAsDFY3zz0ThUbaZ52XS1pCCQ1Df9bQnTgqJNc+u1xQHLYAageKS83uAbtS
nYjBFFuxeRR2FA1n8echCWQV+16Kqq31U1E2yLfWcQKBgQDSoT73pO9h/yN5myqu
XxhAC+Ndas0DTl4pmVPpybpenJerba/0KCfYpcSFHAdkXZ1DYL7U+9uh5NRul09f
WdjayFjn0vv63rwX+PGi1yPHTIv5kLvjYXJtaxzxSzQivYMPmD/7QX4tEsUkpJ8k
90vMSS/C5ieWbpFwWVvEjFbqHQKBgQDNOsTq6USE3R8p4lj9Ror/waUuWVdzZnm3
uZGJw3WzvzaXmqLSfVHntUnD8TPHgk3WZxDhrF73POMADkl9IN/JPI150/Uo6YJo
qYGoZr0mmnEZxVCkwODz5C9icnyjklcRdIRM6eljhFMQDVEacDkptsntHUyIdQZc
L2eLNUfEgwKBgHxy7UNg3lemag110rgIU8mzvHj7m3oymYw2nc/qcwVnvG17d5Tp
DPICr6R+NRfl//9JcDdjQBfdnm5hVHJgIbLS4UTH8j390GDRo+O0/dzJq4KfM4Rb
lUJ1ITqoVnuYQZG7QUJxJd330yedZLJwswZWz7N2TTmixqf9BC2TRd85AoGAN+Qh
bLhKaMSvkACMq61ifXSHP7AlGNB3pYlsEVCh5WnVvEPow9pNTAUbKbmumE7sU8+N
0WfYFQ0H5SP+74zcZTmQbfVDdvjhAw/mt64DJVg6JQKPi87bdJBYNz9mokVgYOiS
fz/Ux71pwZ1e0QxvBOU66NBp31+/c6uVT1wbR3ECgYAdye1+UPpS9Dn89g6Ks0kv
UaFKykXu7vY2uxiNqhmWzze4iq5wmIHmEwc6+rVMluXQPAME7Iya3mBmto9AHQ/n
/ka+fGoaUgAojCLZW5DZcelIETw+Dk+95vyyAUsWfAvn4nKo4/rkBXcSHlvgElzq
SorPiBWYosFB6jqUTXew2w==
-----END PRIVATE KEY-----
`, gun, role))
	}

	return testPEM
}

func TestValidateCertificateWithSHA1(t *testing.T) {
	// Test against SHA1 signature algorithm cert first
	startTime := time.Now()
	template, err := NewCertificate("something", startTime, startTime.AddDate(10, 0, 0))
	require.NoError(t, err)
	// SHA1 signature algorithm is invalid
	template.SignatureAlgorithm = x509.ECDSAWithSHA1
	template.PublicKeyAlgorithm = x509.ECDSA

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	derBytes, err := x509.CreateCertificate(
		rand.Reader, template, template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	sha1Cert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	// Regardless of expiry check, this certificate should fail to validate
	require.Error(t, ValidateCertificate(sha1Cert, false))
	require.Error(t, ValidateCertificate(sha1Cert, true))
}

func TestValidateCertificateWithExpiredCert(t *testing.T) {
	// Test against an expired cert for 10 years ago, only valid for a day
	startTime := time.Now().AddDate(-10, 0, 0)
	template, err := NewCertificate("something", startTime, startTime.AddDate(0, 0, 1))
	require.NoError(t, err)
	template.SignatureAlgorithm = x509.ECDSAWithSHA256
	template.PublicKeyAlgorithm = x509.ECDSA

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	derBytes, err := x509.CreateCertificate(
		rand.Reader, template, template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	expiredCert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	// If we don't check expiry, this cert is perfectly valid
	require.NoError(t, ValidateCertificate(expiredCert, false))
	// We should get an error when we check expiry
	require.Error(t, ValidateCertificate(expiredCert, true))
}

func TestValidateCertificateWithInvalidExpiry(t *testing.T) {
	// Test against a cert with an invalid expiry window: from 10 years in the future to 10 years ago
	startTime := time.Now().AddDate(10, 0, 0)
	template, err := NewCertificate("something", startTime, startTime.AddDate(-10, 0, 0))
	require.NoError(t, err)
	template.SignatureAlgorithm = x509.ECDSAWithSHA256
	template.PublicKeyAlgorithm = x509.ECDSA

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	derBytes, err := x509.CreateCertificate(
		rand.Reader, template, template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	invalidCert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	// Regardless of expiry check, this certificate should fail to validate
	require.Error(t, ValidateCertificate(invalidCert, false))
	require.Error(t, ValidateCertificate(invalidCert, true))
}

func TestValidateCertificateWithShortKey(t *testing.T) {
	startTime := time.Now()
	template, err := NewCertificate("something", startTime, startTime.AddDate(10, 0, 0))
	require.NoError(t, err)
	template.SignatureAlgorithm = x509.SHA256WithRSA
	template.PublicKeyAlgorithm = x509.RSA

	// Use only 1024 bit modulus, this will fail
	weakPrivKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	derBytes, err := x509.CreateCertificate(
		rand.Reader, template, template, &weakPrivKey.PublicKey, weakPrivKey)
	require.NoError(t, err)

	weakKeyCert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	// Regardless of expiry check, this certificate should fail to validate
	require.Error(t, ValidateCertificate(weakKeyCert, false))
	require.Error(t, ValidateCertificate(weakKeyCert, true))
}
