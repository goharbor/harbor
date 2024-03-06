package base

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func (c Credentials) Sign(request *http.Request) *http.Request {
	query := request.URL.Query()

	request.URL.RawQuery = query.Encode()
	return Sign4(request, c)
}

func (c Credentials) SignUrl(request *http.Request) string {
	query := request.URL.Query()
	ldt := timestampV4()
	sdt := ldt[:8]
	meta := new(metadata)
	meta.date, meta.service, meta.region, meta.signedHeaders, meta.algorithm = sdt, c.Service, c.Region, "", "HMAC-SHA256"
	meta.credentialScope = concat("/", meta.date, meta.region, meta.service, "request")

	query.Set("X-Date", ldt)
	query.Set("X-NotSignBody", "")
	query.Set("X-Credential", c.AccessKeyID+"/"+meta.credentialScope)
	query.Set("X-Algorithm", meta.algorithm)
	query.Set("X-SignedHeaders", meta.signedHeaders)
	query.Set("X-SignedQueries", "")
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	query.Set("X-SignedQueries", strings.Join(keys, ";"))

	if c.SessionToken != "" {
		query.Set("X-Security-Token", c.SessionToken)
	}

	// Task 1
	hashedCanonReq := hashedSimpleCanonicalRequestV4(request, query, meta)

	// Task 2
	stringToSign := concat("\n", meta.algorithm, ldt, meta.credentialScope, hashedCanonReq)

	// Task 3
	signingKey := signingKeyV4(c.SecretAccessKey, meta.date, meta.region, meta.service)
	signature := signatureV4(signingKey, stringToSign)

	query.Set("X-Signature", signature)
	return query.Encode()
}

// Sign4 signs a request with Signed Signature Version 4.
func Sign4(request *http.Request, credential Credentials) *http.Request {
	keys := credential

	prepareRequestV4(request)
	meta := new(metadata)
	meta.service, meta.region = keys.Service, keys.Region

	// Task 0 设置SessionToken的header
	if credential.SessionToken != "" {
		request.Header.Set("X-Security-Token", credential.SessionToken)
	}

	// Task 1
	hashedCanonReq := hashedCanonicalRequestV4(request, meta)

	// Task 2
	stringToSign := stringToSignV4(request, hashedCanonReq, meta)

	// Task 3
	signingKey := signingKeyV4(keys.SecretAccessKey, meta.date, meta.region, meta.service)
	signature := signatureV4(signingKey, stringToSign)

	request.Header.Set("Authorization", buildAuthHeaderV4(signature, meta, keys))

	return request
}

func hashedSimpleCanonicalRequestV4(request *http.Request, query url.Values, meta *metadata) string {
	payloadHash := hashSHA256([]byte(""))

	if request.URL.Path == "" {
		request.URL.Path = "/"
	}

	canonicalRequest := concat("\n", request.Method, normuri(request.URL.Path), normquery(query), "\n", meta.signedHeaders, payloadHash)

	return hashSHA256([]byte(canonicalRequest))
}

func hashedCanonicalRequestV4(request *http.Request, meta *metadata) string {
	payload := readAndReplaceBody(request)
	payloadHash := hashSHA256(payload)
	request.Header.Set("X-Content-Sha256", payloadHash)

	request.Header.Set("Host", request.Host)

	var sortedHeaderKeys []string
	for key := range request.Header {
		switch key {
		case "Content-Type", "Content-Md5", "Host", "X-Security-Token":
		default:
			if !strings.HasPrefix(key, "X-") {
				continue
			}
		}
		sortedHeaderKeys = append(sortedHeaderKeys, strings.ToLower(key))
	}
	sort.Strings(sortedHeaderKeys)

	var headersToSign string
	for _, key := range sortedHeaderKeys {
		value := strings.TrimSpace(request.Header.Get(key))
		if key == "host" {
			if strings.Contains(value, ":") {
				split := strings.Split(value, ":")
				port := split[1]
				if port == "80" || port == "443" {
					value = split[0]
				}
			}
		}
		headersToSign += key + ":" + value + "\n"
	}
	meta.signedHeaders = concat(";", sortedHeaderKeys...)
	canonicalRequest := concat("\n", request.Method, normuri(request.URL.Path), normquery(request.URL.Query()), headersToSign, meta.signedHeaders, payloadHash)

	return hashSHA256([]byte(canonicalRequest))
}

func stringToSignV4(request *http.Request, hashedCanonReq string, meta *metadata) string {
	requestTs := request.Header.Get("X-Date")

	meta.algorithm = "HMAC-SHA256"
	meta.date = tsDateV4(requestTs)
	meta.credentialScope = concat("/", meta.date, meta.region, meta.service, "request")

	return concat("\n", meta.algorithm, requestTs, meta.credentialScope, hashedCanonReq)
}

func signatureV4(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
}

func prepareRequestV4(request *http.Request) *http.Request {
	necessaryDefaults := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=utf-8",
		"X-Date":       timestampV4(),
	}

	for header, value := range necessaryDefaults {
		if request.Header.Get(header) == "" {
			request.Header.Set(header, value)
		}
	}

	if request.URL.Path == "" {
		request.URL.Path += "/"
	}

	return request
}

func signingKeyV4(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")
	return kSigning
}

func buildAuthHeaderV4(signature string, meta *metadata, keys Credentials) string {
	credential := keys.AccessKeyID + "/" + meta.credentialScope

	return meta.algorithm +
		" Credential=" + credential +
		", SignedHeaders=" + meta.signedHeaders +
		", Signature=" + signature
}

func timestampV4() string {
	return now().Format(timeFormatV4)
}

func tsDateV4(timestamp string) string {
	return timestamp[:8]
}

func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

func hashSHA256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func hashMD5(content []byte) string {
	h := md5.New()
	h.Write(content)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func readAndReplaceBody(request *http.Request) []byte {
	if request.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(request.Body)
	request.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}

func concat(delim string, str ...string) string {
	return strings.Join(str, delim)
}

var now = func() time.Time {
	return time.Now().UTC()
}

func normuri(uri string) string {
	parts := strings.Split(uri, "/")
	for i := range parts {
		parts[i] = encodePathFrag(parts[i])
	}
	return strings.Join(parts, "/")
}

func encodePathFrag(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}
	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		} else {
			t[j] = c
			j++
		}
	}
	return string(t)
}

func shouldEscape(c byte) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
		return false
	}
	if '0' <= c && c <= '9' {
		return false
	}
	if c == '-' || c == '_' || c == '.' || c == '~' {
		return false
	}
	return true
}

func normquery(v url.Values) string {
	queryString := v.Encode()

	return strings.Replace(queryString, "+", "%20", -1)
}
