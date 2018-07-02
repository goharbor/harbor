package utils

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

func getRSAKey() (data.PrivateKey, error) {
	raw := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAtKGse3BcxXAp5OkLGYq0HfDcCvgag3R/9e8pHUGsJhkSZFrn
ZWAsAVFKSYaYItf1D/g3izqVDMtMpXZ1chNzaRysnbrb/q7JTbiGzXo9FcshyUc9
tcB60wFbvsXE2LaxZcKNxLYXbOvf+tdg/P07oPG24fzYI4+rbZ1wyoORbT1ys33Z
hHyifFvO7rbe69y3HG+xbp7yWYAR4e8Nw9jX8/9sGslAV9vEXOdNL3qlcgsYRGDU
DsUJsnWaMzjstvUxb8mVf9KG2W039ucgaXgBW/jeP3F1VSYFKLd03LvuJ8Ir5E0s
cWjwTd59nm0XbbRI3KiBGnAgrJ4iK07HrUkpDQIDAQABAoIBAHfr1k1lfdH+83Fs
XtgoRAiUviHyMfgQQlwO2eb4kMgCYTmLOJEPVmfRhlZmK18GrUZa7tVaoVYLKum3
SaXg0AB67wcQ5bmiZTdaSPTmMOPlJpsw1wFxtpmcD0MKnfOa5w++KMzub4L63or0
rwmHPi1ODLLgYMbLPW7a1eU9kDFLOnx3RRy9a29hQXxGsRYetrIbKmeDi6c+ndQ8
I5YWObcixxl5GP6CTnEugV7wd2JmXuQRGFdopUwQESCD9VkxDSevQBSPoyZKHxGy
/d3jf0VNlvwsxhD3ybhw8jTN/cmm2LWmP4jylG7iG7YRPVaW/0s39IZ9DnNDwgWB
03Yk2gECgYEA44jcSI5kXOrbBGDdV+wTUoL24Zoc0URX33F15UIOQuQsypaFoRJ6
J23JWEZN99aquuo1+0BBSfiumbpLwSwfXi0nL3oTzS9eOp1HS7AwFGd/OHdpdMsC
w2eInRwCh4GrEf508GXo+tSL2NS8+MFVAG2/SjEf06SroQ/rQ87Qm0ECgYEAyzqr
6YvbAnRBy5GgQlTgxR9r0U8N7lM9g2Tk8uGvYI8zu/Tgia4diHAwK1ymKbl3lf95
3NcHR+ffwOO+nnfFCvmCYXs4ggRCkeopv19bsCLkfnTBNTxPFh6lyLEnn3C+rcCe
ZAkKLrm8BHGviPIgn0aElMQAbhJxTWfClw/VVs0CgYAlDhfZ1R6xJypN9zx04iRv
bpaoPQHubrPk1sR9dpl9+Uz2HTdb+PddznpY3vI5p4Mcd6Ic7eT0GATPUlCeAAKH
wtC74aSx6MHux8hhoiriV8yXNJM/CwTDL+xGsdYTnWFvx8HhmKctmknAIT05QbsH
G9hoS8HEJPAyhbYpz9eXQQKBgQCftPXQTPXJUe86uLBGMEmK34xtKkD6XzPiA/Hf
5PdbXG39cQzbZZcT14YjLWXvOC8AE4qCwACaw1+VR+ROyDRy0W1iieD4W7ysymYQ
XDHDk0gZEEudOE22RlNmCcHnjERsawiN+ISl/5P/sg+OASkdwd8CwZzM43VirP3A
lNLEqQKBgHqj85n8ew23SR0GX0W1PgowaCHHD1p81y2afs1x7H9R1PNuQsTXKnpf
vMiG7Oxakj4WDC5M5KsHWqchqKDycT+J1BfLI58Sf2qo6vtaV+4DARNgzcG/WB4b
VnpsczK/1aUH7iBexuF0YqdPQwzpSvrY0XZcgCFQ52JDn3vjblhX
-----END RSA PRIVATE KEY-----`)
	block, _ := pem.Decode(raw)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privKey, err := RSAToPrivateKey(key)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func TestConvertTUFKeyToPKCS8(t *testing.T) {
	keys := []data.PrivateKey{}

	rsaKey, err := getRSAKey()
	require.NoError(t, err)
	keys = append(keys, rsaKey)

	ecKey, err := GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	keys = append(keys, ecKey)

	edKey, err := GenerateED25519Key(rand.Reader)
	require.NoError(t, err)
	keys = append(keys, edKey)

	for _, k := range keys {
		testConvertKeyToPKCS8(t, k, nil)
		testConvertKeyToPKCS8(t, k, []byte("ponies"))
	}
}

func testConvertKeyToPKCS8(t *testing.T, privKey data.PrivateKey, password []byte) {
	der, err := ConvertTUFKeyToPKCS8(privKey, password)
	require.NoError(t, err, "could not convert private key to pkcs8")

	key, err := ParsePKCS8ToTufKey(der, password)
	require.NoError(t, err, "could not decrypt the newly created pkcs8 key")
	require.EqualValues(t, key.Private(), privKey.Private(), "private key did not match")
}

func TestParsePKCS8ToTufKey(t *testing.T) {
	keys := []struct {
		// algorithm (supports: rsa, ecdsa, ed25519)
		Algorithm string
		// unencrypted PKCS#8 key
		Unencrypted string
		// encrypted PKCS#8 key
		Encrypted string
		// password used to encrypt key
		Password string
	}{
		{
			Algorithm: "rsa",
			Unencrypted: `-----BEGIN PRIVATE KEY-----
MIIEvgIBADALBgkqhkiG9w0BAQEEggSqMIIEpgIBAAKCAQEA3vRQI7s20MF0Zc3f
ywttsw72OkRXuTT0/JQrSuoilzOSaoLKp7sYprIeIu9OeXqvBbwAxe3i1GViGwWM
8cH9QqD05XhMz0Crr9vu2zHaZFEI9mgTXcQxMQGntZ4xYV/rL/fddzj7+n1oKNvo
vS800NvPEMUkkgApdp5ES605V1q51tBpLEYJ82xb5vT8cVseFYfA4G+gVqLNfQla
sa0QsQT4YlVEDbbwT3/wuMG/m+wTx2p8urhD+69oQbORkpqkNiEzMNidOrvtD7qy
ab+cUNamYU0CKOFn/KhWuoZV7EVYnc+oevm7naYsenDq43Q5hGyacEuTjGtLnUG3
2d8RewIDAQABAoIBAQDeLOBfewSY6u8vNAU7tVvP/6znS4uPiHJJ8O1jbgaiXkYd
1dBVbWCXXRAjCA5PiC45rKuokfJkbdNh0houIH5ck0D4GvWP4oY0bRqNXBShuw8P
XY9O9V9/0oJpvga/XnJkDsCnOiX/7FCLxvka7ZvYNfMWZx6WT4sCJZ0xPKHTpT09
SzbhDOCLOsM4nWbs5xkXuEGPkD289z+NOmENdcKDHz0mgYAr7hKJI3oAt2ogTjSy
iQBLmxgudBUP5oJ1qY6/kYUCTYE3cjssY/mqfNcKtylQpTIUaUCw8BhOf3yJFA0G
SI6C7hp96cjEk2dRQxAtYhSZJPA2uN+D1/UIUeSBAoGBAO9VnVaTvkvc55upY9kU
KnZcztJwG2Hf9FRrJrgC2RIaj3KNEImUQyPgIVBXRzvdrvtvNJ6Tdb0cg6CBLJu7
IeQjca2Lj4ACIzoSMF8ak6BP3cdB/fCc2eHchqBKPWgZ23dq3CrpedtR6TbWLcsw
MrYdpZzpZ2eFPeStYxVhTLExAoGBAO56tNX+Sp4N4cCeht6ttljLnGfAjeSBWv4z
+xIqRyEoXbhchoppNxfnX34CrKmQM8MHfEYHKo27u/SkhnMGyuDOtrd51+jhB0LX
COH3w6GI162HVTRJXND8nUtCPB9h/SpFspr8Nk1Y2FtcfwkqhVphzExFzKm7eOPu
eevlsKJrAoGBALuvhh1Y60iOycpWggjAObRsf3yjkbWlbPOuu8Rd52C9F3UbjrZ1
YFmH8FgSubgG1qwyvy8EMLbG36pE4niVvbQs337bDQOzqXBmxywtqUt0llUmOUAx
oOPwjlqxHYq/jE4PrOyx/2+wwpTQTUUkXQBYK4Hrv718zdbA6gzgKsZhAoGBAMsn
QufNMZmFL9xb737Assbf5QRJd1bCj1Zfx7FIzMFFVtlYENDWIrW9R47cDmSAUGgC
923cavbEh7A3e8V/ctKhpeuU40YidIIPFyUQYNo57amI0R+yo1vw5roW2YrOedFK
AIWg901asyzZFeskCufcyiHrkBbDeo+JNtmrGJazAoGBAMOxKBm9HpFXB0h90+FA
6aQgL6FfF578QTW+s5UsZUJjKunYuGmXnYOb3iF6Kvvc/qmKDB1TInVtho1N5TcD
pLTbO3/JPtJyolvYXjnBI7qXjpPxJeicDbA919iDaEVtW1tQOQ9g7WBG0VorWUSr
oQSGi2o00tBJXEiEPmsJK2HL
-----END PRIVATE KEY-----
`,
			Encrypted: `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIFHzBJBgkqhkiG9w0BBQ0wPDAbBgkqhkiG9w0BBQwwDgQIl+qYwi/xfGACAggA
MB0GCWCGSAFlAwQBKgQQT0TeNa0Rap4ngC1Xo5lEKwSCBNDVrsI+c8fmlsPfi98Q
AwIaR62VaolrBjcGvczqFoGMcJs2gcmMOdXBxCQua2E7EAAxnGKUjjkRH4ZyDdaC
hBgEejJ04pXTuhoi9eyuDgi6tVFW+LGmRo0RRQDm5VrB5x1hUFn2EczBnxWoh/Fv
IiaIetdIwGdVwuXyorQzrIqkv7/y2HMuokOXQjWudjK1rNLEi5dce7hgengYGg23
xUiPIPGghkegjguFX/Kkn2+V+RtKAZtjfO6x5gr35EnFZe5FR0zhHPrwwIHm+YCl
4KEJIB3fchgUr8xOIB/WI1YRsubf2mir8SGY1iyePZ3Ya44xian2HpbRPxnHV7xF
FE34p4dzbuWtWBdJCaLWu4sXvPVujf0gxLRRFwskvnMUOWj12ZZvwQtS+hbVcuTC
95J8douCiuM65QOd4uVUdNPQftq6w2F2+V+m9cFUbFuztsCywuDuu7BM6z8ir5JV
GYq5FK5WIWio+EYSUkGhvrwWxFAdwyjTPVKWS8xhyhYfV38q+MpoKT8rbDDi6SbP
XQmtnm26du1C2Esuhd7XsosUBzsT7LENyRQQ/ECtCgJzw/iDGnRPeJW/hGp+Jwot
sWE4MOXbOjHl0kU4RSHwYuZ7AIixeJp7BrOFVQjyUlciRtxK0Qd0dSfeyGT7MRsD
IxU6jOFALGkSeDQGIkVLRUzqVGEWz+rMy7WpxsPqCw6H7mGXPkVgK4G4WRFGlfJo
MpYgSLfnHbh/HL7R5iJSp0GSXd8b3EkOSQXuJM6wJOoXyvnz9UXMeyQcev3fRi7g
XJoj38g410f8+8SIYKUbA629qhbpFn9C3U0pGIqqOBWvkdVYJJf7fxew0eD4ByB1
sZEwRd7iMrH+Iivq4luGK2DpVf0JcyhBOqMTK0PpnIvWFwDTleQ2reyOH4SgfRCm
q+tuECSCWvkmzCcSODwZM/h2A6GtmkdBYoFP0l1KEgInvTG3eniOVYtq4G5SDk1J
nW9JE7egnk56Hg6k9gYwfKwvBOO8l10SO9SEXkD+PXVPMHaaYNhCgdL+s/509TWt
hNvfgplT1A+Q7tMnNYK9D9ZmQvsVhmA2QPhFxJNydvaIggoc5uXDedIc5sX6lIBK
kKUwQaeNoM5JhnG9Dfy0UKoJSLvo43KWeJRkW5guuqlmUcr5blGDwXSXnqKofY+1
XomqtbZf77QstIp0LXG3x2bkxJxERC6UOcZz3mOh46eh52WFOztJ2nKc6v64aiPs
x0QqEkE/7cTT3ntKuvXZliFEMb9sTy5DJ+tm3pOpSC6XaOAFkCjvn9JllF+hCeEh
2jtMBZmpcOKrQpU//q540IbIZtbBR4zrziyMAz1Nhdsx+ziFeRsUeX3cEscQmTgt
3RDskN6xY/y6GBQcJ9sglnZIeAoD3WKsAUHJsbDJ2Cev+HtUpb/Ki/1xnSbw6wpC
5bkcOcVwdPlJ7Cz3nzCYN50YADmZBhGJEi3G+bEH31UdtuCe4/qkNaKsRxitwjxg
JkHt3S7Yw/wvg0WmRfR6cNXUpzqfbldJMfFapKKVNIesrqhZL8JAHNbF+wKFTeSr
Bsy4+8LUtpF26yx4+mTHMEPylWc1s6stNONnMcOCxMHdolbA7isX9JDN6Zr/yAtN
By5Wk9hG6JdPUxl/YkhZUCdKxg==
-----END ENCRYPTED PRIVATE KEY-----
`,
			Password: "ponies",
		},
		{
			Algorithm: "ecdsa",
			Unencrypted: `-----BEGIN PRIVATE KEY-----
MHgCAQAwEAYHKoZIzj0CAQYFK4EEACEEYTBfAgEBBBxdqDSBsFWIAiQ99sRSQZrb
IFczI8UIRM7FD/SNoTwDOgAETbjLZYByEmU3oALoLIz4Xr814S8jMs3cAfJuywm/
kLGZ7y/1i56SXpTOByu6LHXrRokEi4hWQAc=
-----END PRIVATE KEY-----
`,
			Encrypted: `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIHOMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAjsfXVXuwOQnAICCAAw
HQYJYIZIAWUDBAEqBBB0O793rOzupOUavjLSiPmBBIGALJxsXCe8rLBfeviStIH0
A+1jCXUqXNm8D4npyNui/JRi/CjYPqgcO/2ulP8ppUAeTnLVQdhpv5ZOemK5ibMc
ECaNuzo40snnpve4duZEufkI9hXrO6MAMRT+G5ep1rKyIKboIPkzYUAdezj5ggUu
p1Gc8HB7j2SYjQX0Ybvlr6k=
-----END ENCRYPTED PRIVATE KEY-----
`,
			Password: "ponies",
		},
		{
			Algorithm: "ecdsa",
			Unencrypted: `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgiwt5YfD/xQdVwJZ0
2TpiJDQQ8DRHYVeWzIscya52BvChRANCAAT58IHVQJwbo3/MS/dFjh+xM85gVydX
xY+wxYDkaougZDPIgvu3+bQZ4xYSAnCGX7UJIiLloKuuuvbmXQlnSGqw
-----END PRIVATE KEY-----
`,
			Encrypted: `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIHeMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAhTsVpOdfLrSwICCAAw
HQYJYIZIAWUDBAEqBBDjOBeXbqjRH1FP6BkI6n3PBIGQgdYzQ7wOKaEp73WloPJl
966A0tiBCt2wy4LSueFjlh5NtF0o8odzg+zK9lHGD0MluWwM9LsDk5xtfcXCE6Ya
16PfcoAKE6l7VuGob4wDms/Y0G9DLhKXxErQOfzEolNZjN5RcZF9938ZPjQUDeIX
yYeijYPrkZWmdcrFPZUkKY5HQMIjXoULlUtlN9fFckLn
-----END ENCRYPTED PRIVATE KEY-----
`,
			Password: "ponies",
		},
		{
			Algorithm: "ecdsa",
			Unencrypted: `-----BEGIN PRIVATE KEY-----
MIG2AgEAMBAGByqGSM49AgEGBSuBBAAiBIGeMIGbAgEBBDCnjVESo9F+BLL4ZSt1
/ZU14MYlozCa7OyjdcdFjwMSajUZ4N0HVoBpJoeFh8DKaJ2hZANiAAQ4sTZRVUFU
p4IXBI9QEuwWh0Lsd/uUtZkpwXrjC4hpCQI3am7QC5Ct83VAtQ1WXBYg7EjIYNfi
CDbvJdq1y0IhdY138OQvsTaewiuYHUvRwjljxiSjpNEOB6AoD36FlqY=
-----END PRIVATE KEY-----
`,
			Encrypted: `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIBDjBJBgkqhkiG9w0BBQ0wPDAbBgkqhkiG9w0BBQwwDgQIPvZXgw/gzOsCAggA
MB0GCWCGSAFlAwQBKgQQIEKz02wepEM1NmHfAHd76wSBwALvEtH7pz03/m6Z5Gkv
aafc7pfImJJzpLVIxGcNxrLkz1/WFoIpXHR8Bdde4dBEa8TYz91KvSNfnFjGE2xk
AiMSZyuObwGB0Vw1de8tDlpsVYftkZC4VrpRwUEksTUzYgHum/sqRlm7DmeXJq2t
540HZ9XhS2ZfT1bSqaCMX1s98KMpxDDHRDPX0SEBskyGqIWKLzLfYJnf07OlZ8r4
/oByTKigO+k9U40jNeuYW1aZeM8wqdApa89K48jWxftfNQ==
-----END ENCRYPTED PRIVATE KEY-----
`,
			Password: "ponies",
		},
		{
			Algorithm: "ecdsa",
			Unencrypted: `-----BEGIN PRIVATE KEY-----
MIHuAgEAMBAGByqGSM49AgEGBSuBBAAjBIHWMIHTAgEBBEIB0dZtwbNAy6K2iJF0
P9cTcwv2XnSCyeiIcOW/IG3I09pklXQNCw1igQdKSjZLZZRVS4OZMvuG774OPq9j
F7m/tkihgYkDgYYABADN4kHmO0/+mIHmIuC6id/lX04mZ9wZovU102l4VUdZA3e6
tZWDMdS2D3oqwhud2xCoHNw2ShxspzUISd/srH1pPAA3L2r2eZ6axrEqz1unbdBy
q1SyrsbtvDEJsP8STxiK3RSL9r00gqwlK44lp6dYQU3zd6IzS/69ACj/nmfX+YE4
AA==
-----END PRIVATE KEY-----
`,
			Encrypted: `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIBTzBJBgkqhkiG9w0BBQ0wPDAbBgkqhkiG9w0BBQwwDgQI0ykvTci7+gkCAggA
MB0GCWCGSAFlAwQBKgQQdd7n4DwrOqViFfmLvN7ERASCAQCuGSdBA8+YCSPpVhcj
bYxIU0knqQPrlF5N3+BJMGVIz3468DVZFi9UtiiKRaHGfSxaUimqyQ0oUzXULEav
ZSp6abjxoBJZXPJWHu0f85s3DOjoks990a5o+J72gH/dH9yK/GgvR3MSXlkKoMut
Zqm3toQAF5RReU3E2wirnYEec8h4Zw5gy3FX63MuvX9mhlOtHTPhiZjoM2ogVQ3b
iH1BRu4nZF11wSZNxtWflLGMGZaQv3zBt4w7eu6AN25U5DiTy6ReWmsyB+kD/cZ5
eR9Noh2UlGMDPqTTrv2xqjjo/ieTusS7aGVQ0d8VESsxaAmCYx/kDY3FLtrKvsb5
kB1E
-----END ENCRYPTED PRIVATE KEY-----
`,
			Password: "ponies",
		},
		{
			Algorithm: "ed25519",
			Unencrypted: `-----BEGIN PRIVATE KEY-----
MHICAQAwCwYJKwYBBAHaRw8BBGDkASR4b08nd+A8txI3h+1hG+7EAIxE5cdbv3gt
rwib9ibygTpRt8XjscMv+vum4zFjI2pPZbhQn6lZlumHo7g35AEkeG9PJ3fgPLcS
N4ftYRvuxACMROXHW794La8Im/Y=
-----END PRIVATE KEY-----
`,
			Encrypted: `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIHOMEkGCSqGSIb3DQEFDTA8MBsGCSqGSIb3DQEFDDAOBAhdzEpEYPy6ugICCAAw
HQYJYIZIAWUDBAEqBBCBSlYxcs3xzXsdF3JXvyvuBIGA9EaQcxFFk6d6jQiTzACC
66TFviduVfuqCr+VmulBAQZiCLj3PCpKugQ5z0aJ9CPfCRus5II3qS+qOXjI3OuH
hmevc2qO2C9bpsDteibfi9/tJJ8vVHdd+w44rSbLdLFro+p9CTT1R/VQillIFT4N
lXM12s3lyKsXT8bUcibd0gM=
-----END ENCRYPTED PRIVATE KEY-----
`,
			Password: "ponies",
		},
	}

	for _, k := range keys {
		testParsePKCS8ToTufKey(t, k.Algorithm, []byte(k.Unencrypted), []byte(k.Encrypted), []byte(k.Password))
	}
}

func testParsePKCS8ToTufKey(t *testing.T, alg string, unencypted, encrypted, password []byte) {
	block, _ := pem.Decode(unencypted)
	require.NotNil(t, block.Bytes, "could not decode PEM block properly")

	unencryptedKey, err := ParsePKCS8ToTufKey(block.Bytes, nil)
	require.NoError(t, err, "could not parse pkcs8 to tuf key")
	require.Equal(t, alg, unencryptedKey.Algorithm())

	block, _ = pem.Decode(encrypted)
	require.NotNil(t, block.Bytes, "could not decode PEM block properly")

	encryptedKey, err := ParsePKCS8ToTufKey(block.Bytes, password)
	require.NoError(t, err, "could not parse pkcs8 to tuf key")
	require.Equal(t, alg, encryptedKey.Algorithm())
	require.EqualValues(t, unencryptedKey.Private(), encryptedKey.Private())

	_, err = ParsePKCS8ToTufKey(block.Bytes, []byte("wrong password"))
	require.Error(t, err, "could parse key even with wrong password")
}

func TestPEMtoPEM(t *testing.T) {
	testInputPKCS1 := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEA3vRQI7s20MF0Zc3fywttsw72OkRXuTT0/JQrSuoilzOSaoLK
p7sYprIeIu9OeXqvBbwAxe3i1GViGwWM8cH9QqD05XhMz0Crr9vu2zHaZFEI9mgT
XcQxMQGntZ4xYV/rL/fddzj7+n1oKNvovS800NvPEMUkkgApdp5ES605V1q51tBp
LEYJ82xb5vT8cVseFYfA4G+gVqLNfQlasa0QsQT4YlVEDbbwT3/wuMG/m+wTx2p8
urhD+69oQbORkpqkNiEzMNidOrvtD7qyab+cUNamYU0CKOFn/KhWuoZV7EVYnc+o
evm7naYsenDq43Q5hGyacEuTjGtLnUG32d8RewIDAQABAoIBAQDeLOBfewSY6u8v
NAU7tVvP/6znS4uPiHJJ8O1jbgaiXkYd1dBVbWCXXRAjCA5PiC45rKuokfJkbdNh
0houIH5ck0D4GvWP4oY0bRqNXBShuw8PXY9O9V9/0oJpvga/XnJkDsCnOiX/7FCL
xvka7ZvYNfMWZx6WT4sCJZ0xPKHTpT09SzbhDOCLOsM4nWbs5xkXuEGPkD289z+N
OmENdcKDHz0mgYAr7hKJI3oAt2ogTjSyiQBLmxgudBUP5oJ1qY6/kYUCTYE3cjss
Y/mqfNcKtylQpTIUaUCw8BhOf3yJFA0GSI6C7hp96cjEk2dRQxAtYhSZJPA2uN+D
1/UIUeSBAoGBAO9VnVaTvkvc55upY9kUKnZcztJwG2Hf9FRrJrgC2RIaj3KNEImU
QyPgIVBXRzvdrvtvNJ6Tdb0cg6CBLJu7IeQjca2Lj4ACIzoSMF8ak6BP3cdB/fCc
2eHchqBKPWgZ23dq3CrpedtR6TbWLcswMrYdpZzpZ2eFPeStYxVhTLExAoGBAO56
tNX+Sp4N4cCeht6ttljLnGfAjeSBWv4z+xIqRyEoXbhchoppNxfnX34CrKmQM8MH
fEYHKo27u/SkhnMGyuDOtrd51+jhB0LXCOH3w6GI162HVTRJXND8nUtCPB9h/SpF
spr8Nk1Y2FtcfwkqhVphzExFzKm7eOPueevlsKJrAoGBALuvhh1Y60iOycpWggjA
ObRsf3yjkbWlbPOuu8Rd52C9F3UbjrZ1YFmH8FgSubgG1qwyvy8EMLbG36pE4niV
vbQs337bDQOzqXBmxywtqUt0llUmOUAxoOPwjlqxHYq/jE4PrOyx/2+wwpTQTUUk
XQBYK4Hrv718zdbA6gzgKsZhAoGBAMsnQufNMZmFL9xb737Assbf5QRJd1bCj1Zf
x7FIzMFFVtlYENDWIrW9R47cDmSAUGgC923cavbEh7A3e8V/ctKhpeuU40YidIIP
FyUQYNo57amI0R+yo1vw5roW2YrOedFKAIWg901asyzZFeskCufcyiHrkBbDeo+J
NtmrGJazAoGBAMOxKBm9HpFXB0h90+FA6aQgL6FfF578QTW+s5UsZUJjKunYuGmX
nYOb3iF6Kvvc/qmKDB1TInVtho1N5TcDpLTbO3/JPtJyolvYXjnBI7qXjpPxJeic
DbA919iDaEVtW1tQOQ9g7WBG0VorWUSroQSGi2o00tBJXEiEPmsJK2HL
-----END RSA PRIVATE KEY-----
`)

	testOutputPKCS8 := []byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADALBgkqhkiG9w0BAQEEggSqMIIEpgIBAAKCAQEA3vRQI7s20MF0Zc3f
ywttsw72OkRXuTT0/JQrSuoilzOSaoLKp7sYprIeIu9OeXqvBbwAxe3i1GViGwWM
8cH9QqD05XhMz0Crr9vu2zHaZFEI9mgTXcQxMQGntZ4xYV/rL/fddzj7+n1oKNvo
vS800NvPEMUkkgApdp5ES605V1q51tBpLEYJ82xb5vT8cVseFYfA4G+gVqLNfQla
sa0QsQT4YlVEDbbwT3/wuMG/m+wTx2p8urhD+69oQbORkpqkNiEzMNidOrvtD7qy
ab+cUNamYU0CKOFn/KhWuoZV7EVYnc+oevm7naYsenDq43Q5hGyacEuTjGtLnUG3
2d8RewIDAQABAoIBAQDeLOBfewSY6u8vNAU7tVvP/6znS4uPiHJJ8O1jbgaiXkYd
1dBVbWCXXRAjCA5PiC45rKuokfJkbdNh0houIH5ck0D4GvWP4oY0bRqNXBShuw8P
XY9O9V9/0oJpvga/XnJkDsCnOiX/7FCLxvka7ZvYNfMWZx6WT4sCJZ0xPKHTpT09
SzbhDOCLOsM4nWbs5xkXuEGPkD289z+NOmENdcKDHz0mgYAr7hKJI3oAt2ogTjSy
iQBLmxgudBUP5oJ1qY6/kYUCTYE3cjssY/mqfNcKtylQpTIUaUCw8BhOf3yJFA0G
SI6C7hp96cjEk2dRQxAtYhSZJPA2uN+D1/UIUeSBAoGBAO9VnVaTvkvc55upY9kU
KnZcztJwG2Hf9FRrJrgC2RIaj3KNEImUQyPgIVBXRzvdrvtvNJ6Tdb0cg6CBLJu7
IeQjca2Lj4ACIzoSMF8ak6BP3cdB/fCc2eHchqBKPWgZ23dq3CrpedtR6TbWLcsw
MrYdpZzpZ2eFPeStYxVhTLExAoGBAO56tNX+Sp4N4cCeht6ttljLnGfAjeSBWv4z
+xIqRyEoXbhchoppNxfnX34CrKmQM8MHfEYHKo27u/SkhnMGyuDOtrd51+jhB0LX
COH3w6GI162HVTRJXND8nUtCPB9h/SpFspr8Nk1Y2FtcfwkqhVphzExFzKm7eOPu
eevlsKJrAoGBALuvhh1Y60iOycpWggjAObRsf3yjkbWlbPOuu8Rd52C9F3UbjrZ1
YFmH8FgSubgG1qwyvy8EMLbG36pE4niVvbQs337bDQOzqXBmxywtqUt0llUmOUAx
oOPwjlqxHYq/jE4PrOyx/2+wwpTQTUUkXQBYK4Hrv718zdbA6gzgKsZhAoGBAMsn
QufNMZmFL9xb737Assbf5QRJd1bCj1Zfx7FIzMFFVtlYENDWIrW9R47cDmSAUGgC
923cavbEh7A3e8V/ctKhpeuU40YidIIPFyUQYNo57amI0R+yo1vw5roW2YrOedFK
AIWg901asyzZFeskCufcyiHrkBbDeo+JNtmrGJazAoGBAMOxKBm9HpFXB0h90+FA
6aQgL6FfF578QTW+s5UsZUJjKunYuGmXnYOb3iF6Kvvc/qmKDB1TInVtho1N5TcD
pLTbO3/JPtJyolvYXjnBI7qXjpPxJeicDbA919iDaEVtW1tQOQ9g7WBG0VorWUSr
oQSGi2o00tBJXEiEPmsJK2HL
-----END PRIVATE KEY-----
`)

	block, _ := pem.Decode(testInputPKCS1)
	require.NotEmpty(t, block)

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	testPrivKey, err := RSAToPrivateKey(rsaKey)
	require.NoError(t, err)

	der, err := ConvertTUFKeyToPKCS8(testPrivKey, nil)
	require.NoError(t, err, "could not convert pkcs1 to pkcs8")

	testOutput := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	require.EqualValues(t, testOutputPKCS8, testOutput)
}
