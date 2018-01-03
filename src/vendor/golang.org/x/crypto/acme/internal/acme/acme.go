// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package acme provides an implementation of the
// Automatic Certificate Management Environment (ACME) spec.
// See https://tools.ietf.org/html/draft-ietf-acme-acme-02 for details.
//
// This package is a work in progress and makes no API stability promises.
package acme

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// LetsEncryptURL is the Directory endpoint of Let's Encrypt CA.
const LetsEncryptURL = "https://acme-v01.api.letsencrypt.org/directory"

// Client is an ACME client.
// The only required field is Key. An example of creating a client with a new key
// is as follows:
//
// 	key, err := rsa.GenerateKey(rand.Reader, 2048)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	client := &Client{Key: key}
//
type Client struct {
	// HTTPClient optionally specifies an HTTP client to use
	// instead of http.DefaultClient.
	HTTPClient *http.Client

	// Key is the account key used to register with a CA and sign requests.
	// Key.Public() must return a *rsa.PublicKey or *ecdsa.PublicKey.
	Key crypto.Signer

	// DirectoryURL points to the CA directory endpoint.
	// If empty, LetsEncryptURL is used.
	// Mutating this value after a successful call of Client's Discover method
	// will have no effect.
	DirectoryURL string

	dirMu sync.Mutex // guards writes to dir
	dir   *Directory // cached result of Client's Discover method
}

// Discover performs ACME server discovery using c.DirectoryURL.
//
// It caches successful result. So, subsequent calls will not result in
// a network round-trip. This also means mutating c.DirectoryURL after successful call
// of this method will have no effect.
func (c *Client) Discover() (Directory, error) {
	c.dirMu.Lock()
	defer c.dirMu.Unlock()
	if c.dir != nil {
		return *c.dir, nil
	}

	dirURL := c.DirectoryURL
	if dirURL == "" {
		dirURL = LetsEncryptURL
	}
	res, err := c.httpClient().Get(dirURL)
	if err != nil {
		return Directory{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Directory{}, responseError(res)
	}

	var v struct {
		Reg    string `json:"new-reg"`
		Authz  string `json:"new-authz"`
		Cert   string `json:"new-cert"`
		Revoke string `json:"revoke-cert"`
		Meta   struct {
			Terms   string   `json:"terms-of-service"`
			Website string   `json:"website"`
			CAA     []string `json:"caa-identities"`
		}
	}
	if json.NewDecoder(res.Body).Decode(&v); err != nil {
		return Directory{}, err
	}
	c.dir = &Directory{
		RegURL:    v.Reg,
		AuthzURL:  v.Authz,
		CertURL:   v.Cert,
		RevokeURL: v.Revoke,
		Terms:     v.Meta.Terms,
		Website:   v.Meta.Website,
		CAA:       v.Meta.CAA,
	}
	return *c.dir, nil
}

// CreateCert requests a new certificate.
// In the case where CA server does not provide the issued certificate in the response,
// CreateCert will poll certURL using c.FetchCert, which will result in additional round-trips.
// In such scenario the caller can cancel the polling with ctx.
//
// If the bundle is true, the returned value will also contain CA (the issuer) certificate.
// The csr is a DER encoded certificate signing request.
func (c *Client) CreateCert(ctx context.Context, csr []byte, exp time.Duration, bundle bool) (der [][]byte, certURL string, err error) {
	if _, err := c.Discover(); err != nil {
		return nil, "", err
	}

	req := struct {
		Resource  string `json:"resource"`
		CSR       string `json:"csr"`
		NotBefore string `json:"notBefore,omitempty"`
		NotAfter  string `json:"notAfter,omitempty"`
	}{
		Resource: "new-cert",
		CSR:      base64.RawURLEncoding.EncodeToString(csr),
	}
	now := timeNow()
	req.NotBefore = now.Format(time.RFC3339)
	if exp > 0 {
		req.NotAfter = now.Add(exp).Format(time.RFC3339)
	}

	res, err := c.postJWS(c.dir.CertURL, req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return nil, "", responseError(res)
	}

	curl := res.Header.Get("location") // cert permanent URL
	if res.ContentLength == 0 {
		// no cert in the body; poll until we get it
		cert, err := c.FetchCert(ctx, curl, bundle)
		return cert, curl, err
	}
	// slurp issued cert and ca, if requested
	cert, err := responseCert(c.httpClient(), res, bundle)
	return cert, curl, err
}

// FetchCert retrieves already issued certificate from the given url, in DER format.
// It retries the request until the certificate is successfully retrieved,
// context is cancelled by the caller or an error response is received.
//
// The returned value will also contain CA (the issuer) certificate if bundle is true.
func (c *Client) FetchCert(ctx context.Context, url string, bundle bool) ([][]byte, error) {
	for {
		res, err := c.httpClient().Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusOK {
			return responseCert(c.httpClient(), res, bundle)
		}
		if res.StatusCode > 299 {
			return nil, responseError(res)
		}
		d, err := retryAfter(res.Header.Get("retry-after"))
		if err != nil {
			d = 3 * time.Second
		}
		select {
		case <-time.After(d):
			// retry
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// AcceptTOS always returns true to indicate the acceptance of a CA Terms of Service
// during account registration. See Register method of Client for more details.
func AcceptTOS(string) bool { return true }

// Register creates a new account registration by following the "new-reg" flow.
// It returns registered account. The a argument is not modified.
//
// The registration may require the caller to agree to the CA Terms of Service (TOS).
// If so, and the account has not indicated the acceptance of the terms (see Account for details),
// Register calls prompt with a TOS URL provided by the CA. Prompt should report
// whether the caller agrees to the terms. To always accept the terms, the caller can use AcceptTOS.
func (c *Client) Register(a *Account, prompt func(tos string) bool) (*Account, error) {
	if _, err := c.Discover(); err != nil {
		return nil, err
	}

	var err error
	if a, err = c.doReg(c.dir.RegURL, "new-reg", a); err != nil {
		return nil, err
	}
	var accept bool
	if a.CurrentTerms != "" && a.CurrentTerms != a.AgreedTerms {
		accept = prompt(a.CurrentTerms)
	}
	if accept {
		a.AgreedTerms = a.CurrentTerms
		a, err = c.UpdateReg(a)
	}
	return a, err
}

// GetReg retrieves an existing registration.
// The url argument is an Account URI.
func (c *Client) GetReg(url string) (*Account, error) {
	a := &Account{URI: url}
	return c.doReg(url, "reg", a)
}

// UpdateReg updates an existing registration.
// It returns an updated account copy. The provided account is not modified.
func (c *Client) UpdateReg(a *Account) (*Account, error) {
	return c.doReg(a.URI, "reg", a)
}

// Authorize performs the initial step in an authorization flow.
// The caller will then need to choose from and perform a set of returned
// challenges using c.Accept in order to successfully complete authorization.
func (c *Client) Authorize(domain string) (*Authorization, error) {
	if _, err := c.Discover(); err != nil {
		return nil, err
	}

	type authzID struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	req := struct {
		Resource   string  `json:"resource"`
		Identifier authzID `json:"identifier"`
	}{
		Resource:   "new-authz",
		Identifier: authzID{Type: "dns", Value: domain},
	}
	res, err := c.postJWS(c.dir.AuthzURL, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return nil, responseError(res)
	}

	var v wireAuthz
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}
	if v.Status != StatusPending {
		return nil, fmt.Errorf("Unexpected status: %s", v.Status)
	}
	return v.authorization(res.Header.Get("Location")), nil
}

// GetAuthz retrieves the current status of an authorization flow.
//
// A client typically polls an authz status using this method.
func (c *Client) GetAuthz(url string) (*Authorization, error) {
	res, err := c.httpClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, responseError(res)
	}
	var v wireAuthz
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}
	return v.authorization(url), nil
}

// GetChallenge retrieves the current status of an challenge.
//
// A client typically polls a challenge status using this method.
func (c *Client) GetChallenge(url string) (*Challenge, error) {
	res, err := c.httpClient().Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, responseError(res)
	}
	v := wireChallenge{URI: url}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}
	return v.challenge(), nil
}

// Accept informs the server that the client accepts one of its challenges
// previously obtained with c.Authorize.
//
// The server will then perform the validation asynchronously.
func (c *Client) Accept(chal *Challenge) (*Challenge, error) {
	auth, err := keyAuth(c.Key.Public(), chal.Token)
	if err != nil {
		return nil, err
	}

	req := struct {
		Resource string `json:"resource"`
		Type     string `json:"type"`
		Auth     string `json:"keyAuthorization"`
	}{
		Resource: "challenge",
		Type:     chal.Type,
		Auth:     auth,
	}
	res, err := c.postJWS(chal.URI, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	// Note: the protocol specifies 200 as the expected response code, but
	// letsencrypt seems to be returning 202.
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		return nil, responseError(res)
	}

	var v wireChallenge
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}
	return v.challenge(), nil
}

// HTTP01Handler creates a new handler which responds to a http-01 challenge.
// The token argument is a Challenge.Token value.
func (c *Client) HTTP01Handler(token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, token) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("content-type", "text/plain")
		auth, err := keyAuth(c.Key.Public(), token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(auth))
	})
}

// TLSSNI01ChallengeCert creates a certificate for TLS-SNI-01 challenge response.
// Servers can present the certificate to validate the challenge and prove control
// over a domain name.
//
// The implementation is incomplete in that the returned value is a single certificate,
// computed only for Z0 of the key authorization. ACME CAs are expected to update
// their implementations to use the newer version, TLS-SNI-02.
// For more details on TLS-SNI-01 see https://tools.ietf.org/html/draft-ietf-acme-acme-01#section-7.3.
//
// The token argument is a Challenge.Token value.
// The returned certificate is valid for the next 24 hours.
func (c *Client) TLSSNI01ChallengeCert(token string) (tls.Certificate, error) {
	ka, err := keyAuth(c.Key.Public(), token)
	if err != nil {
		return tls.Certificate{}, nil
	}
	b := sha256.Sum256([]byte(ka))
	h := hex.EncodeToString(b[:])
	name := fmt.Sprintf("%s.%s.acme.invalid", h[:32], h[32:])
	return tlsChallengeCert(name)
}

// TLSSNI02ChallengeCert creates a certificate for TLS-SNI-02 challenge response.
// Servers can present the certificate to validate the challenge and prove control
// over a domain name. For more details on TLS-SNI-02 see
// https://tools.ietf.org/html/draft-ietf-acme-acme-03#section-7.3.
//
// The token argument is a Challenge.Token value.
// The returned certificate is valid for the next 24 hours.
func (c *Client) TLSSNI02ChallengeCert(token string) (tls.Certificate, error) {
	b := sha256.Sum256([]byte(token))
	h := hex.EncodeToString(b[:])
	sanA := fmt.Sprintf("%s.%s.token.acme.invalid", h[:32], h[32:])

	ka, err := keyAuth(c.Key.Public(), token)
	if err != nil {
		return tls.Certificate{}, nil
	}
	b = sha256.Sum256([]byte(ka))
	h = hex.EncodeToString(b[:])
	sanB := fmt.Sprintf("%s.%s.ka.acme.invalid", h[:32], h[32:])

	return tlsChallengeCert(sanA, sanB)
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// postJWS signs body and posts it to the provided url.
// The body argument must be JSON-serializable.
func (c *Client) postJWS(url string, body interface{}) (*http.Response, error) {
	nonce, err := fetchNonce(c.httpClient(), url)
	if err != nil {
		return nil, err
	}
	b, err := jwsEncodeJSON(body, c.Key, nonce)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	return c.httpClient().Do(req)
}

// doReg sends all types of registration requests.
// The type of request is identified by typ argument, which is a "resource"
// in the ACME spec terms.
//
// A non-nil acct argument indicates whether the intention is to mutate data
// of the Account. Only Contact and Agreement of its fields are used
// in such cases.
//
// The fields of acct will be populate with the server response
// and may be overwritten.
func (c *Client) doReg(url string, typ string, acct *Account) (*Account, error) {
	req := struct {
		Resource  string   `json:"resource"`
		Contact   []string `json:"contact,omitempty"`
		Agreement string   `json:"agreement,omitempty"`
	}{
		Resource: typ,
	}
	if acct != nil {
		req.Contact = acct.Contact
		req.Agreement = acct.AgreedTerms
	}
	res, err := c.postJWS(url, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, responseError(res)
	}

	var v struct {
		Contact        []string
		Agreement      string
		Authorizations string
		Certificates   string
	}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("Decode: %v", err)
	}
	return &Account{
		URI:            res.Header.Get("Location"),
		Contact:        v.Contact,
		AgreedTerms:    v.Agreement,
		CurrentTerms:   linkHeader(res.Header, "terms-of-service"),
		Authz:          linkHeader(res.Header, "next"),
		Authorizations: v.Authorizations,
		Certificates:   v.Certificates,
	}, nil
}

func responseCert(client *http.Client, res *http.Response, bundle bool) ([][]byte, error) {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %v", err)
	}
	cert := [][]byte{b}
	if !bundle {
		return cert, nil
	}

	// append ca cert
	up := linkHeader(res.Header, "up")
	if up == "" {
		return nil, errors.New("rel=up link not found")
	}
	res, err = client.Get(up)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, responseError(res)
	}
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return append(cert, b), nil
}

// responseError creates an error of Error type from resp.
func responseError(resp *http.Response) error {
	// don't care if ReadAll returns an error:
	// json.Unmarshal will fail in that case anyway
	b, _ := ioutil.ReadAll(resp.Body)
	e := struct {
		Status int
		Type   string
		Detail string
	}{
		Status: resp.StatusCode,
	}
	if err := json.Unmarshal(b, &e); err != nil {
		// this is not a regular error response:
		// populate detail with anything we received,
		// e.Status will already contain HTTP response code value
		e.Detail = string(b)
		if e.Detail == "" {
			e.Detail = resp.Status
		}
	}
	return &Error{
		StatusCode:  e.Status,
		ProblemType: e.Type,
		Detail:      e.Detail,
		Header:      resp.Header,
	}
}

func fetchNonce(client *http.Client, url string) (string, error) {
	resp, err := client.Head(url)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()
	enc := resp.Header.Get("replay-nonce")
	if enc == "" {
		return "", errors.New("nonce not found")
	}
	return enc, nil
}

func linkHeader(h http.Header, rel string) string {
	for _, v := range h["Link"] {
		parts := strings.Split(v, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if !strings.HasPrefix(p, "rel=") {
				continue
			}
			if v := strings.Trim(p[4:], `"`); v == rel {
				return strings.Trim(parts[0], "<>")
			}
		}
	}
	return ""
}

func retryAfter(v string) (time.Duration, error) {
	if i, err := strconv.Atoi(v); err == nil {
		return time.Duration(i) * time.Second, nil
	}
	t, err := http.ParseTime(v)
	if err != nil {
		return 0, err
	}
	return t.Sub(timeNow()), nil
}

// keyAuth generates a key authorization string for a given token.
func keyAuth(pub crypto.PublicKey, token string) (string, error) {
	th, err := JWKThumbprint(pub)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", token, th), nil
}

// tlsChallengeCert creates a temporary certificate for TLS-SNI challenges
// with the given SANs.
func tlsChallengeCert(san ...string) (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	t := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageKeyEncipherment,
		DNSNames:              san,
	}
	der, err := x509.CreateCertificate(rand.Reader, &t, &t, &key.PublicKey, key)
	return tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  key,
	}, nil
}

// encodePEM returns b encoded as PEM with block of type typ.
func encodePEM(typ string, b []byte) []byte {
	pb := &pem.Block{Type: typ, Bytes: b}
	return pem.EncodeToMemory(pb)
}

// timeNow is useful for testing for fixed current time.
var timeNow = time.Now
