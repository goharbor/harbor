package signed

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"testing"
	"text/template"

	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

type KeyTemplate struct {
	KeyType string
}

const baseRSAKey = `{"keytype":"{{.KeyType}}","keyval":{"public":"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyyvBtTg2xzYS+MTTIBqSpI4V78tt8Yzqi7Jki/Z6NqjiDvcnbgcTqNR2t6B2W5NjGdp/hSaT2jyHM+kdmEGaPxg/zIuHbL3NIp4e0qwovWiEgACPIaELdn8O/kt5swsSKl1KMvLCH1sM86qMibNMAZ/hXOwd90TcHXCgZ91wHEAmsdjDC3dB0TT+FBgOac8RM01Y196QrZoOaDMTWh0EQfw7YbXAElhFVDFxBzDdYWbcIHSIogXQmq0CP+zaL/1WgcZZIClt2M6WCaxxF1S34wNn45gCvVZiZQ/iKWHerSr/2dGQeGo+7ezMSutRzvJ+01fInD86RS/CEtBCFZ1VyQIDAQAB","private":"MIIEpAIBAAKCAQEAyyvBtTg2xzYS+MTTIBqSpI4V78tt8Yzqi7Jki/Z6NqjiDvcnbgcTqNR2t6B2W5NjGdp/hSaT2jyHM+kdmEGaPxg/zIuHbL3NIp4e0qwovWiEgACPIaELdn8O/kt5swsSKl1KMvLCH1sM86qMibNMAZ/hXOwd90TcHXCgZ91wHEAmsdjDC3dB0TT+FBgOac8RM01Y196QrZoOaDMTWh0EQfw7YbXAElhFVDFxBzDdYWbcIHSIogXQmq0CP+zaL/1WgcZZIClt2M6WCaxxF1S34wNn45gCvVZiZQ/iKWHerSr/2dGQeGo+7ezMSutRzvJ+01fInD86RS/CEtBCFZ1VyQIDAQABAoIBAHar8FFxrE1gAGTeUpOF8fG8LIQMRwO4U6eVY7V9GpWiv6gOJTHXYFxU/aL0Ty3eQRxwy9tyVRo8EJz5pRex+e6ws1M+jLOviYqW4VocxQ8dZYd+zBvQfWmRfah7XXJ/HPUx2I05zrmR7VbGX6Bu4g5w3KnyIO61gfyQNKF2bm2Q3yblfupx3URvX0bl180R/+QN2Aslr4zxULFE6b+qJqBydrztq+AAP3WmskRxGa6irFnKxkspJqUpQN1mFselj6iQrzAcwkRPoCw0RwCCMq1/OOYvQtgxTJcO4zDVlbw54PvnxPZtcCWw7fO8oZ2Fvo2SDo75CDOATOGaT4Y9iqECgYEAzWZSpFbN9ZHmvq1lJQg//jFAyjsXRNn/nSvyLQILXltz6EHatImnXo3v+SivG91tfzBI1GfDvGUGaJpvKHoomB+qmhd8KIQhO5MBdAKZMf9fZqZofOPTD9xRXECCwdi+XqHBmL+l1OWz+O9Bh+Qobs2as/hQVgHaoXhQpE0NkTcCgYEA/Tjf6JBGl1+WxQDoGZDJrXoejzG9OFW19RjMdmPrg3t4fnbDtqTpZtCzXxPTCSeMrvplKbqAqZglWyq227ksKw4p7O6YfyhdtvC58oJmivlLr6sFaTsER7mDcYce8sQpqm+XQ8IPbnOk0Z1l6g56euTwTnew49uy25M6U1xL0P8CgYEAxEXv2Kw+OVhHV5PX4BBHHj6we88FiDyMfwM8cvfOJ0datekf9X7ImZkmZEAVPJpWBMD+B0J0jzU2b4SLjfFVkzBHVOH2Ob0xCH2MWPAWtekin7OKizUlPbW5ZV8b0+Kq30DQ/4a7D3rEhK8UPqeuX1tHZox1MAqrgbq3zJj4yvcCgYEAktYPKPm4pYCdmgFrlZ+bA0iEPf7Wvbsd91F5BtHsOOM5PQQ7e0bnvWIaEXEad/2CG9lBHlBy2WVLjDEZthILpa/h6e11ao8KwNGY0iKBuebT17rxOVMqqTjPGt8CuD2994IcEgOPFTpkAdUmyvG4XlkxbB8F6St17NPUB5DGuhsCgYA//Lfytk0FflXEeRQ16LT1YXgV7pcR2jsha4+4O5pxSFw/kTsOfJaYHg8StmROoyFnyE3sg76dCgLn0LENRCe5BvDhJnp5bMpQldG3XwcAxH8FGFNY4LtV/2ZKnJhxcONkfmzQPOmTyedOzrKQ+bNURsqLukCypP7/by6afBY4dA=="}}`
const baseRSAx509Key = `{"keytype":"{{.KeyType}}","keyval":{"public":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZLekNDQXhXZ0F3SUJBZ0lRVERENFNHZS8yaUVJQjAxZFhzTHE5akFMQmdrcWhraUc5dzBCQVFzd09ERWEKTUJnR0ExVUVDaE1SWkc5amEyVnlMbU52YlM5dWIzUmhjbmt4R2pBWUJnTlZCQU1URVdSdlkydGxjaTVqYjIwdgpibTkwWVhKNU1CNFhEVEUxTURjeE56QXdNemt6TVZvWERURTNNRGN4TmpBd016a3pNVm93T0RFYU1CZ0dBMVVFCkNoTVJaRzlqYTJWeUxtTnZiUzl1YjNSaGNua3hHakFZQmdOVkJBTVRFV1J2WTJ0bGNpNWpiMjB2Ym05MFlYSjUKTUlJQ0lqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FnOEFNSUlDQ2dLQ0FnRUFzYnY1R01oS21DK0J6V3hhZTZSTQpyaHVxU082VkNpb1dhMkZmdmtzYlhtaDhNaktTeUFHQUJLSnVoVksyTHI3NWsrZktTeVVOSFpEWUYxVXFhMnljCnlDQndVVVFXYTNqVDdPaFI3T2FzRDBXYVJEL2MvODhkVlNRejVsdEV0Y3IvVGpRSHpqcjVXL2dWWXJIVkV2UXkKWUJGS3BkSHdRRGpLNnZ6Njc1WnRxMjBKUStzcXRNNFlUeis5dXg5Y25LQTFuc0JDM1YvVk1ybTRWZ3pqL0lOawppL0ZNMVh0Yjd1UFFTK0hRSElYc2R5SktsTHdsTXg2RjhTeFpYZHROUjh4bE8wdkI3Qm5KK0hKZlBWTEpYZDVmCjRld1lUZkE3WmtRNWJESzAzeDRkSzloR2VFUjlEVURja2tNM3RVZHhleTFQZXBkK1BSWWRsL0k5MG5UYXN0L2EKdmpQdUxkYjR2KzFuVnZQVzhjMHNvaGhQK1VkUUU1UHA0dlhDUmMvbVZ1S1NNbEU5bFRDYzY2TFoxcm5tQzJ4agpzKzNpcWZWeWFIcjFmbnUrZjdhTk9jSmFSTlpRN1ErWEdBbXljMXJ6MmJzTmc0S1pIVGdHQnBzNHMwQWV6MUhSCll6NW94QmVUVEZUaUNtZFBJS2lhbGhrTmJQWS9FUXQzNzBXYjIzMTVUQm12QkI0NitXbWoycG9ka2FJeGhJLzMKblRwT25mZlFub25rTmRwdTlKeFJWNFlWQTh1b0hvNWc1WjMreEF1S1A1TnlRSHdvQkwwbElKbTdBK2lHUDF5cwpuNnFVVk5ab0dENFR4ZHduYllmMFBKa3ZhL0FjLzkvaktyajVyQ2NrL1NDOVllbGNxaCt1dklENDA5YU9veVAzCk83SDF0bmQ4M1hXZm5nczRuaFNmQmdzQ0F3RUFBYU0xTURNd0RnWURWUjBQQVFIL0JBUURBZ0NnTUJNR0ExVWQKSlFRTU1Bb0dDQ3NHQVFVRkJ3TURNQXdHQTFVZEV3RUIvd1FDTUFBd0N3WUpLb1pJaHZjTkFRRUxBNElDQVFCawpnRGExQ0F2NWY4QWFaN1YvN2JoMzBtUnZEajVwNGtPWkpSclc3UDdsTTJZU2tNVWFnQzB1QTQySUU1TjI4R2VWCkp4VEh6V0FCQThwNE1ZREllTVZJdGJndHBUWVorZTByL3RPL3pnRHVxNEZ5UFZmZllRYlh1L2k5N29TbVNQYy8KRXpZQktrNHl1RHZ3ZjZtNjJPSGxNalZNcitzM3pQUHB4dFFTaFRndkJ4QWp2ekdmVFBRSEZSdm5jWFZPd2dyRAp0ampsS3RzMGx0azI4eWJ3dyt3SVVCdWg0dzNrZFVBR2RYME9sY3NIdnM3TFhoc01XcmdxMUs4ZVNJZlR6YUdGClMwcE5MNEZObUV4VDVKaFk1SnZ4cWRxclB2RFJEU2FOUXV0OHc2K2FpeXVPVFpiZDRQeTlLZHd2bUNrNk5GdHoKd3lpWUwzT2hZa201Ui9iUm93YVY1dWwrY1BETmV0cGV3WnZJQTUzUkJYZlZCejl0TXI0M2ZaaW9YRFltNTkyVQpKTE1GaGRWMm1zYk9McWFIcGRoN0JhWFFITGxEdHZpaUVLdVRqalJKWEZWTk9seTA1UHBxeFhjWnRSbHhpRjhhCkoveWJ5a1Y0aWc0K3U1aVNGK2dFZjkyaWpaRTNjNnlsYkZjSDhoRVV0bTRqSElHZ1JsWGJ3NmZvV3llb2Z5VUIKTk5COTZyVG5UdkxmdDlReGprUjdlNGgycU41MnFIOVY5L3NLSjlSVFFqU1RERXM3MDF2Z1ZVd0tpVC9VZ3hLTAp3UzJ5dnZJeTN5TFpFUGltQnF6emFSeStCZ3Q4anNrNnQvNEdIT2Y0Rzk0a3paMkIyNUJnYjV5MTl2WVdDQSswCitXdlRCeGdxb0o1Y2lCdXMxYWJiUjZORU1RbXQyeUtneTZEejNJVXgxZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K","private":"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"}}`
const baseECDSAKey = `
{"keytype":"{{.KeyType}}","keyval":{"public":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEgl3rzMPMEKhS1k/AX16MM4PdidpjJr+z4pj0Td+30QnpbOIARgpyR1PiFztU8BZlqG3cUazvFclr2q/xHvfrqw==","private":"MHcCAQEEIDqtcdzU7H3AbIPSQaxHl9+xYECt7NpK7B1+6ep5cv9CoAoGCCqGSM49AwEHoUQDQgAEgl3rzMPMEKhS1k/AX16MM4PdidpjJr+z4pj0Td+30QnpbOIARgpyR1PiFztU8BZlqG3cUazvFclr2q/xHvfrqw=="}}`
const baseECDSAx509Key = `{"keytype":"ecdsa-x509","keyval":{"public":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJwRENDQVVtZ0F3SUJBZ0lRQlBWc1NoUmRMdG45WEtRZ29JaDIvREFLQmdncWhrak9QUVFEQWpBNE1Sb3cKR0FZRFZRUUtFeEZrYjJOclpYSXVZMjl0TDI1dmRHRnllVEVhTUJnR0ExVUVBeE1SWkc5amEyVnlMbU52YlM5dQpiM1JoY25rd0hoY05NVFV3TnpFek1EVXdORFF4V2hjTk1UY3dOekV5TURVd05EUXhXakE0TVJvd0dBWURWUVFLCkV4RmtiMk5yWlhJdVkyOXRMMjV2ZEdGeWVURWFNQmdHQTFVRUF4TVJaRzlqYTJWeUxtTnZiUzl1YjNSaGNua3cKV1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVI3SjNSOGpWODV5Rnp0dGFTV3FMRDFHa042UHlhWAowUUdmOHh2Rzd6MUYwUG5DQUdSWk9QQ01aWWpZSGVkdzNXY0FmQWVVcDY5OVExSjNEYW9kbzNBcm96VXdNekFPCkJnTlZIUThCQWY4RUJBTUNBS0F3RXdZRFZSMGxCQXd3Q2dZSUt3WUJCUVVIQXdNd0RBWURWUjBUQVFIL0JBSXcKQURBS0JnZ3Foa2pPUFFRREFnTkpBREJHQWlFQWppVkJjaTBDRTBaazgwZ2ZqbytYdE9xM3NURGJkSWJRRTZBTQpoL29mN1RFQ0lRRGxlbXB5MDRhY0RKODNnVHBvaFNtcFJYdjdJbnRLc0lRTU1oLy9VZzliU2c9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==","private":null}}`

func TestRSAPSSVerifier(t *testing.T) {
	// Unmarshal our private RSA Key
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.RSAKey})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Sign some data using RSAPSS
	message := []byte("test data for signing")
	hash := crypto.SHA256
	hashed := sha256.Sum256(message)
	signedData, err := rsaPSSSign(testRSAKey, hash, hashed[:])
	require.NoError(t, err)

	// Create and call Verify on the verifier
	rsaVerifier := RSAPSSVerifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.NoError(t, err, "expecting success but got error while verifying data using RSA PSS")
}

func TestRSAPSSx509Verifier(t *testing.T) {
	// Unmarshal our public RSA Key
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAx509Key)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.RSAx509Key})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signed message
	signedData, _ := hex.DecodeString("3de02fa54cdba45c67860f058b7cff1ba264610dc3c5b466b7df027bc52068bdf2956fe438dba08b0b71daa0780a3037bf8f50a09d91ca81fa872bbdbbbff6ef17e04df8741ad5c2f2c3ea5de97d6ffaf4999c83fdfba4b6cb2443da11c7b7eea84123c2fdaf3319fa6342cbbdbd1aa25d1ac20aeee687e48cbf191cc8f68049230261469eeada33dec0af74287766bd984dd01820a7edfb8b0d030e2fcf00886c578b07eb905a2eebc81fd982a578e717c7ac773cab345950c71e1eaf81b70401e5bf3c67cdcb9068bf4b50ff0456b530b3cec5586827eb39b123f9d666a65f4b418a355438ed1753da8a27577ab9cd791d7b840c7e34ecc1290c46d98aa0dd73c0427f6ef8f63e36af42e9657520b8f56c9231ba7e0172dfc3456c63c54e9eae95d06bafe571e91afa1e42d4010e60dd5c441df112cc8474253eee7f1d6c5350039ffcd1f8b0bb013e4403c16fc5b40d6bd56b742ea1ed82c87880147db194b33b022077cc2e8d31ef3eada3e46683ad437ad8ef7ecbe03c29d7a53a9771e42cc4f9d782813c491186fde2cd1dfa408c4e21dd4c3ca1664e901772ffe1713e37b07c9287572114865a05e17cbe29d8622c6b033dcb43c9721d0943c58098607cc28bd58b3caf3dfc1f66d01ebfaf1aa5c2c5945c23af83fe114e587fa7bcbaea6bdccff3c0ad03ce3328f67af30168e225e5827ad9e94b4702de984e6dd775")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	rsaVerifier := RSAPSSVerifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.NoError(t, err, "expecting success but got error while verifying data using RSAPSS and an X509 encoded Key")
}

func TestRSAPSSVerifierWithInvalidKeyType(t *testing.T) {
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: "rsa-invalid"})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signed data with invalidRsaKeyJSON
	signedData, _ := hex.DecodeString("2741a57a5ef89f841b4e0a6afbcd7940bc982cd919fbd11dfc21b5ccfe13855b9c401e3df22da5480cef2fa585d0f6dfc6c35592ed92a2a18001362c3a17f74da3906684f9d81c5846bf6a09e2ede6c009ae164f504e6184e666adb14eadf5f6e12e07ff9af9ad49bf1ea9bcfa3bebb2e33be7d4c0fabfe39534f98f1e3c4bff44f637cff3dae8288aea54d86476a3f1320adc39008eae24b991c1de20744a7967d2e685ac0bcc0bc725947f01c9192ffd3e9300eba4b7faa826e84478493fdf97c705dd331dd46072050d6c5e317c2d63df21694dbaf909ebf46ce0ff04f3979fe13723ae1a823c65f27e56efa19e88f9e7b8ee56eac34353b944067deded3a")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	rsaVerifier := RSAPSSVerifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.Error(t, err, "invalid key type for RSAPSS verifier: rsa-invalid")
}

func TestRSAPSSVerifierWithInvalidKeyLength(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	err = verifyPSS(key.Public(), nil, nil)
	require.Error(t, err)
	require.IsType(t, ErrInvalidKeyLength{}, err)
}

func TestRSAPSSVerifierWithInvalidKey(t *testing.T) {
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseECDSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: "ecdsa"})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signed data with invalidRsaKeyJSON
	signedData, _ := hex.DecodeString("2741a57a5ef89f841b4e0a6afbcd7940bc982cd919fbd11dfc21b5ccfe13855b9c401e3df22da5480cef2fa585d0f6dfc6c35592ed92a2a18001362c3a17f74da3906684f9d81c5846bf6a09e2ede6c009ae164f504e6184e666adb14eadf5f6e12e07ff9af9ad49bf1ea9bcfa3bebb2e33be7d4c0fabfe39534f98f1e3c4bff44f637cff3dae8288aea54d86476a3f1320adc39008eae24b991c1de20744a7967d2e685ac0bcc0bc725947f01c9192ffd3e9300eba4b7faa826e84478493fdf97c705dd331dd46072050d6c5e317c2d63df21694dbaf909ebf46ce0ff04f3979fe13723ae1a823c65f27e56efa19e88f9e7b8ee56eac34353b944067deded3a")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	rsaVerifier := RSAPSSVerifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.Error(t, err, "invalid key type for RSAPSS verifier: ecdsa")
}

func TestRSAPSSVerifierWithInvalidSignature(t *testing.T) {
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.RSAKey})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Sign some data using RSAPSS
	message := []byte("test data for signing")
	hash := crypto.SHA256
	hashed := sha256.Sum256(message)
	signedData, err := rsaPSSSign(testRSAKey, hash, hashed[:])
	require.NoError(t, err)

	// Modify the signature
	signedData[0]++

	// Create and call Verify on the verifier
	rsaVerifier := RSAPSSVerifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.Error(t, err, "signature verification failed")
}

func TestRSAPKCS1v15Verifier(t *testing.T) {
	// Unmarshal our private RSA Key
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.RSAKey})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Sign some data using RSAPKCS1v15
	message := []byte("test data for signing")
	hash := crypto.SHA256
	hashed := sha256.Sum256(message)
	signedData, err := rsaPKCS1v15Sign(testRSAKey, hash, hashed[:])
	require.NoError(t, err)

	// Create and call Verify on the verifier
	rsaVerifier := RSAPKCS1v15Verifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.NoError(t, err, "expecting success but got error while verifying data using RSAPKCS1v15")
}

func TestRSAPKCS1v15x509Verifier(t *testing.T) {
	// Unmarshal our public RSA Key
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAx509Key)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.RSAx509Key})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signed message
	signedData, _ := hex.DecodeString("a19602f609646d57f3d0db930bbe491a997baf33f13191916713734ae778ddb4898ece2078741bb0c24d726514c6b4538c3665c374b0b8ec9ff234b45459633268224c9962756ad3684aca5f13a286657375e798ddcb857ed2707c900f097666b958df56b43b790357430c2e7a5c379ba9972c8b008363c144aac5c7e0fbfad83cf6855cf73baf8e3ad774e910ba6ac8dc4cce58fe19cffb7b0a1feaa73d23ebd2d59de2d7d9e98a809d73a310c5396df64ff7a22d735e661e39d37a6c4a013caa6005e91f597ea35db24e6c750d704d292a180128dcf72a818c53a96b0a83ba0414a3611097905262eb79a6ced1484af27c7da6809aa21ae7c6f05ae6568d5e5d9c170470213a30caf2340c3d52e7bd4056d22074daffee6e29d0a6fd3ca6dbd001831fb1e48573f3663b63e110cde19efaf56e49a835aeda82e4d7286de591376ecd03de36d402ec703f39f79b2f764f991d8950a119f2618f6d4e4618114900597a1e89ced609949410623a17b97095afe08babc4c295ade954f055ca01b7909f5585e98eb99bd916583476aa877d20da8f4fe35c0867e934f41c935d469664b80904a93f9f4d9432cabd9383e08559d6452f8e12b2d861412c450709ff874ad63c25a640605a41c4073f0eb4e16e1965abf8e088e210cbf9d3ca884ec2c13fc8a288cfcef2425d9607fcab01dab45c5c346671a9ae1d0e52c81379fa212c")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	rsaVerifier := RSAPKCS1v15Verifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.NoError(t, err, "expecting success but got error while verifying data using RSAPKCS1v15 and an X509 encoded Key")
}

func TestRSAPKCS1v15VerifierWithInvalidKeyType(t *testing.T) {
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: "rsa-invalid"})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signed data with invalidRsaKeyJSON
	signedData, _ := hex.DecodeString("2741a57a5ef89f841b4e0a6afbcd7940bc982cd919fbd11dfc21b5ccfe13855b9c401e3df22da5480cef2fa585d0f6dfc6c35592ed92a2a18001362c3a17f74da3906684f9d81c5846bf6a09e2ede6c009ae164f504e6184e666adb14eadf5f6e12e07ff9af9ad49bf1ea9bcfa3bebb2e33be7d4c0fabfe39534f98f1e3c4bff44f637cff3dae8288aea54d86476a3f1320adc39008eae24b991c1de20744a7967d2e685ac0bcc0bc725947f01c9192ffd3e9300eba4b7faa826e84478493fdf97c705dd331dd46072050d6c5e317c2d63df21694dbaf909ebf46ce0ff04f3979fe13723ae1a823c65f27e56efa19e88f9e7b8ee56eac34353b944067deded3a")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	rsaVerifier := RSAPKCS1v15Verifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.Error(t, err, "invalid key type for RSAPKCS1v15 verifier: rsa-invalid")
}

func TestRSAPKCS1v15VerifierWithInvalidKey(t *testing.T) {
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseECDSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: "ecdsa"})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signed data with invalidRsaKeyJSON
	signedData, _ := hex.DecodeString("2741a57a5ef89f841b4e0a6afbcd7940bc982cd919fbd11dfc21b5ccfe13855b9c401e3df22da5480cef2fa585d0f6dfc6c35592ed92a2a18001362c3a17f74da3906684f9d81c5846bf6a09e2ede6c009ae164f504e6184e666adb14eadf5f6e12e07ff9af9ad49bf1ea9bcfa3bebb2e33be7d4c0fabfe39534f98f1e3c4bff44f637cff3dae8288aea54d86476a3f1320adc39008eae24b991c1de20744a7967d2e685ac0bcc0bc725947f01c9192ffd3e9300eba4b7faa826e84478493fdf97c705dd331dd46072050d6c5e317c2d63df21694dbaf909ebf46ce0ff04f3979fe13723ae1a823c65f27e56efa19e88f9e7b8ee56eac34353b944067deded3a")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	rsaVerifier := RSAPKCS1v15Verifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.Error(t, err, "invalid key type for RSAPKCS1v15 verifier: ecdsa")
}

func TestRSAPKCS1v15VerifierWithInvalidSignature(t *testing.T) {
	var testRSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.RSAKey})

	testRSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Sign some data using RSAPKCS1v15
	message := []byte("test data for signing")
	hash := crypto.SHA256
	hashed := sha256.Sum256(message)
	signedData, err := rsaPKCS1v15Sign(testRSAKey, hash, hashed[:])
	require.NoError(t, err)

	// Modify the signature
	signedData[0]++

	// Create and call Verify on the verifier
	rsaVerifier := RSAPKCS1v15Verifier{}
	err = rsaVerifier.Verify(testRSAKey, signedData, message)
	require.Error(t, err, "signature verification failed")
}

func TestECDSAVerifier(t *testing.T) {
	var testECDSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseECDSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.ECDSAKey})

	testECDSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Sign some data using ECDSA
	message := []byte("test data for signing")
	hashed := sha256.Sum256(message)
	signedData, err := ecdsaSign(testECDSAKey, hashed[:])
	require.NoError(t, err)

	// Create and call Verify on the verifier
	ecdsaVerifier := ECDSAVerifier{}
	err = ecdsaVerifier.Verify(testECDSAKey, signedData, message)
	require.NoError(t, err, "expecting success but got error while verifying data using ECDSA")
}

func TestECDSAVerifierOtherCurves(t *testing.T) {
	curves := []elliptic.Curve{elliptic.P256(), elliptic.P384(), elliptic.P521()}

	for _, curve := range curves {
		ecdsaPrivKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		require.NoError(t, err)

		// Get a DER-encoded representation of the PublicKey
		ecdsaPubBytes, err := x509.MarshalPKIXPublicKey(&ecdsaPrivKey.PublicKey)
		require.NoError(t, err, "failed to marshal public key")

		// Get a DER-encoded representation of the PrivateKey
		ecdsaPrivKeyBytes, err := x509.MarshalECPrivateKey(ecdsaPrivKey)
		require.NoError(t, err, "failed to marshal private key")

		testECDSAPubKey := data.NewECDSAPublicKey(ecdsaPubBytes)
		testECDSAKey, err := data.NewECDSAPrivateKey(testECDSAPubKey, ecdsaPrivKeyBytes)
		require.NoError(t, err, "failed to read private key")

		// Sign some data using ECDSA
		message := []byte("test data for signing")
		hashed := sha256.Sum256(message)
		signedData, err := ecdsaSign(testECDSAKey, hashed[:])
		require.NoError(t, err)

		// Create and call Verify on the verifier
		ecdsaVerifier := ECDSAVerifier{}
		err = ecdsaVerifier.Verify(testECDSAKey, signedData, message)
		require.NoError(t, err, "expecting success but got error while verifying data using ECDSA")

		// Make sure an invalid signature fails verification
		signedData[0]++
		err = ecdsaVerifier.Verify(testECDSAKey, signedData, message)
		require.Error(t, err, "expecting error but got success while verifying data using ECDSA")
	}
}

func TestECDSAx509Verifier(t *testing.T) {
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseECDSAx509Key)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.ECDSAx509Key})

	testECDSAKey, err := data.UnmarshalPublicKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signature for message
	signedData, _ := hex.DecodeString("b82e0ed5c5dddd74c8d3602bfd900c423511697c3cfe54e1d56b9c1df599695c53aa0caafcdc40df3ef496d78ccf67750ba9413f1ccbd8b0ef137f0da1ee9889")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	ecdsaVerifier := ECDSAVerifier{}
	err = ecdsaVerifier.Verify(testECDSAKey, signedData, message)
	require.NoError(t, err, "expecting success but got error while verifying data using ECDSA and an x509 encoded key")
}

func TestECDSAVerifierWithInvalidKeyType(t *testing.T) {
	var testECDSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseECDSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: "ecdsa-invalid"})

	testECDSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signature using invalidECDSAx509Key
	signedData, _ := hex.DecodeString("7b1c45a4dd488a087db46ee459192d890d4f52352620cb84c2c10e0ce8a67fd6826936463a91ffdffab8e6f962da6fc3d3e5735412f7cd161a9fcf97ba1a7033")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	ecdsaVerifier := ECDSAVerifier{}
	err = ecdsaVerifier.Verify(testECDSAKey, signedData, message)
	require.Error(t, err, "invalid key type for ECDSA verifier: ecdsa-invalid")
}

func TestECDSAVerifierWithInvalidKey(t *testing.T) {
	var testECDSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseRSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: "rsa"})

	testECDSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Valid signature using invalidECDSAx509Key
	signedData, _ := hex.DecodeString("7b1c45a4dd488a087db46ee459192d890d4f52352620cb84c2c10e0ce8a67fd6826936463a91ffdffab8e6f962da6fc3d3e5735412f7cd161a9fcf97ba1a7033")
	message := []byte("test data for signing")

	// Create and call Verify on the verifier
	ecdsaVerifier := ECDSAVerifier{}
	err = ecdsaVerifier.Verify(testECDSAKey, signedData, message)
	require.Error(t, err, "invalid key type for ECDSA verifier: rsa")
}

func TestECDSAVerifierWithInvalidSignature(t *testing.T) {
	var testECDSAKey data.PrivateKey
	var jsonKey bytes.Buffer

	// Execute our template
	templ, _ := template.New("KeyTemplate").Parse(baseECDSAKey)
	templ.Execute(&jsonKey, KeyTemplate{KeyType: data.ECDSAKey})

	testECDSAKey, err := data.UnmarshalPrivateKey(jsonKey.Bytes())
	require.NoError(t, err)

	// Sign some data using ECDSA
	message := []byte("test data for signing")
	hashed := sha256.Sum256(message)
	signedData, err := ecdsaSign(testECDSAKey, hashed[:])
	require.NoError(t, err)

	// Modify the signature
	signedData[0]++

	// Create and call Verify on the verifier
	ecdsaVerifier := ECDSAVerifier{}
	err = ecdsaVerifier.Verify(testECDSAKey, signedData, message)
	require.Error(t, err, "signature verification failed")

}

func TestED25519VerifierInvalidKeyType(t *testing.T) {
	key := data.NewPublicKey("bad_type", nil)
	v := Ed25519Verifier{}
	err := v.Verify(key, nil, nil)
	require.Error(t, err)
	require.IsType(t, ErrInvalidKeyType{}, err)
}

func TestRSAPyCryptoVerifierInvalidKeyType(t *testing.T) {
	key := data.NewPublicKey("bad_type", nil)
	v := RSAPyCryptoVerifier{}
	err := v.Verify(key, nil, nil)
	require.Error(t, err)
	require.IsType(t, ErrInvalidKeyType{}, err)
}

func TestPyCryptoRSAPSSCompat(t *testing.T) {
	pubPem := "-----BEGIN PUBLIC KEY-----\nMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAnKuXZeefa2LmgxaL5NsM\nzKOHNe+x/nL6ik+lDBCTV6OdcwAhHQS+PONGhrChIUVR6Vth3hUCrreLzPO73Oo5\nVSCuRJ53UronENl6lsa5mFKP8StYLvIDITNvkoT3j52BJIjyNUK9UKY9As2TNqDf\nBEPIRp28ev/NViwGOEkBu2UAbwCIdnDXm8JQErCZA0Ydm7PKGgjLbFsFGrVzqXHK\n6pdzJXlhr9yap3UpgQ/iO9JtoEYB2EXsnSrPc9JRjR30bNHHtnVql3fvinXrAEwq\n3xmN4p+R4VGzfdQN+8Kl/IPjqWB535twhFYEG/B7Ze8IwbygBjK3co/KnOPqMUrM\nBI8ztvPiogz+MvXb8WvarZ6TMTh8ifZI96r7zzqyzjR1hJulEy3IsMGvz8XS2J0X\n7sXoaqszEtXdq5ef5zKVxkiyIQZcbPgmpHLq4MgfdryuVVc/RPASoRIXG4lKaTJj\n1ANMFPxDQpHudCLxwCzjCb+sVa20HBRPTnzo8LSZkI6jAgMBAAE=\n-----END PUBLIC KEY-----"
	testStr := "The quick brown fox jumps over the lazy dog."
	sigHex := "4e05ee9e435653549ac4eddbc43e1a6868636e8ea6dbec2564435afcb0de47e0824cddbd88776ddb20728c53ecc90b5d543d5c37575fda8bd0317025fc07de62ee8084b1a75203b1a23d1ef4ac285da3d1fc63317d5b2cf1aafa3e522acedd366ccd5fe4a7f02a42922237426ca3dc154c57408638b9bfaf0d0213855d4e9ee621db204151bcb13d4dbb18f930ec601469c992c84b14e9e0b6f91ac9517bb3b749dd117e1cbac2e4acb0e549f44558a2005898a226d5b6c8b9291d7abae0d9e0a16858b89662a085f74a202deb867acab792bdbd2c36731217caea8b17bd210c29b890472f11e5afdd1dd7b69004db070e04201778f2c49f5758643881403d45a58d08f51b5c63910c6185892f0b590f191d760b669eff2464456f130239bba94acf54a0cb98f6939ff84ae26a37f9b890be259d9b5d636f6eb367b53e895227d7d79a3a88afd6d28c198ee80f6527437c5fbf63accb81709925c4e03d1c9eaee86f58e4bd1c669d6af042dbd412de0d13b98b1111e2fadbe34b45de52125e9a"
	k := data.NewPublicKey(data.RSAKey, []byte(pubPem))

	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		t.Fatal(err)
	}
	v := RSAPyCryptoVerifier{}
	err = v.Verify(k, sigBytes, []byte(testStr))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPyNaCled25519Compat(t *testing.T) {
	pubHex := "846612b43cef909a0e4ea9c818379bca4723a2020619f95e7a0ccc6f0850b7dc"
	testStr := "The quick brown fox jumps over the lazy dog."
	sigHex := "166e7013e48f26dccb4e68fe4cf558d1cd3af902f8395534336a7f8b4c56588694aa3ac671767246298a59d5ef4224f02c854f41bfcfe70241db4be1546d6a00"

	pub, _ := hex.DecodeString(pubHex)
	k := data.NewPublicKey(data.ED25519Key, pub)

	sigBytes, _ := hex.DecodeString(sigHex)

	err := Verifiers[data.EDDSASignature].Verify(k, sigBytes, []byte(testStr))
	if err != nil {
		t.Fatal(err)
	}
}

func rsaPSSSign(privKey data.PrivateKey, hash crypto.Hash, hashed []byte) ([]byte, error) {
	if privKey, ok := privKey.(*data.RSAPrivateKey); !ok {
		return nil, fmt.Errorf("private key type not supported: %s", privKey.Algorithm())
	}

	// Create an rsa.PrivateKey out of the private key bytes
	rsaPrivKey, err := x509.ParsePKCS1PrivateKey(privKey.Private())
	if err != nil {
		return nil, err
	}

	// Use the RSA key to RSASSA-PSS sign the data
	sig, err := rsa.SignPSS(rand.Reader, rsaPrivKey, hash, hashed[:], &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash})
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func rsaPKCS1v15Sign(privKey data.PrivateKey, hash crypto.Hash, hashed []byte) ([]byte, error) {
	if privKey, ok := privKey.(*data.RSAPrivateKey); !ok {
		return nil, fmt.Errorf("private key type not supported: %s", privKey.Algorithm())
	}

	// Create an rsa.PrivateKey out of the private key bytes
	rsaPrivKey, err := x509.ParsePKCS1PrivateKey(privKey.Private())
	if err != nil {
		return nil, err
	}

	// Use the RSA key to RSAPKCS1v15 sign the data
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivKey, hash, hashed[:])
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func ecdsaSign(privKey data.PrivateKey, hashed []byte) ([]byte, error) {
	if _, ok := privKey.(*data.ECDSAPrivateKey); !ok {
		return nil, fmt.Errorf("private key type not supported: %s", privKey.Algorithm())
	}

	// Create an ecdsa.PrivateKey out of the private key bytes
	ecdsaPrivKey, err := x509.ParseECPrivateKey(privKey.Private())
	if err != nil {
		return nil, err
	}

	// Use the ECDSA key to sign the data
	r, s, err := ecdsa.Sign(rand.Reader, ecdsaPrivKey, hashed[:])
	if err != nil {
		return nil, err
	}

	rBytes, sBytes := r.Bytes(), s.Bytes()
	octetLength := (ecdsaPrivKey.Params().BitSize + 7) >> 3

	// MUST include leading zeros in the output
	rBuf := make([]byte, octetLength-len(rBytes), octetLength)
	sBuf := make([]byte, octetLength-len(sBytes), octetLength)

	rBuf = append(rBuf, rBytes...)
	sBuf = append(sBuf, sBytes...)

	return append(rBuf, sBuf...), nil
}
