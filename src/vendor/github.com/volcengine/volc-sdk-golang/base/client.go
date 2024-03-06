package base

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	accessKey = "VOLC_ACCESSKEY"
	secretKey = "VOLC_SECRETKEY"

	defaultScheme = "http"
)

var _GlobalClient *http.Client

func init() {
	_GlobalClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     10 * time.Second,
		},
	}
}

// Client 基础客户端
type Client struct {
	Client      *http.Client
	SdkVersion  string
	ServiceInfo *ServiceInfo
	ApiInfoList map[string]*ApiInfo
}

// NewClient 生成一个客户端
func NewClient(info *ServiceInfo, apiInfoList map[string]*ApiInfo) *Client {
	client := &Client{Client: _GlobalClient, ServiceInfo: info.Clone(), ApiInfoList: apiInfoList}

	if client.ServiceInfo.Scheme == "" {
		client.ServiceInfo.Scheme = defaultScheme
	}

	if os.Getenv(accessKey) != "" && os.Getenv(secretKey) != "" {
		client.ServiceInfo.Credentials.AccessKeyID = os.Getenv(accessKey)
		client.ServiceInfo.Credentials.SecretAccessKey = os.Getenv(secretKey)
	} else if _, err := os.Stat(os.Getenv("HOME") + "/.volc/config"); err == nil {
		if content, err := ioutil.ReadFile(os.Getenv("HOME") + "/.volc/config"); err == nil {
			m := make(map[string]string)
			json.Unmarshal(content, &m)
			if accessKey, ok := m["ak"]; ok {
				client.ServiceInfo.Credentials.AccessKeyID = accessKey
			}
			if secretKey, ok := m["sk"]; ok {
				client.ServiceInfo.Credentials.SecretAccessKey = secretKey
			}
		}
	}

	content, err := ioutil.ReadFile("VERSION")
	if err == nil {
		client.SdkVersion = strings.TrimSpace(string(content))
		client.ServiceInfo.Header.Set("User-Agent", strings.Join([]string{"volc-sdk-golang", client.SdkVersion}, "/"))
	}

	return client
}

func (serviceInfo *ServiceInfo) Clone() *ServiceInfo {
	ret := new(ServiceInfo)
	//base info
	ret.Timeout = serviceInfo.Timeout
	ret.Host = serviceInfo.Host
	ret.Scheme = serviceInfo.Scheme

	//credential
	ret.Credentials = serviceInfo.Credentials.Clone()

	// header
	ret.Header = serviceInfo.Header.Clone()
	return ret
}

func (cred Credentials) Clone() Credentials {
	return Credentials{
		Service:         cred.Service,
		Region:          cred.Region,
		SecretAccessKey: cred.SecretAccessKey,
		AccessKeyID:     cred.AccessKeyID,
		SessionToken:    cred.SessionToken,
	}
}

// SetAccessKey 设置AK
func (client *Client) SetAccessKey(ak string) {
	if ak != "" {
		client.ServiceInfo.Credentials.AccessKeyID = ak
	}
}

// SetSecretKey 设置SK
func (client *Client) SetSecretKey(sk string) {
	if sk != "" {
		client.ServiceInfo.Credentials.SecretAccessKey = sk
	}
}

// SetSessionToken
func (client *Client) SetSessionToken(token string) {
	if token != "" {
		client.ServiceInfo.Credentials.SessionToken = token
	}
}

// SetHost 设置Host
func (client *Client) SetHost(host string) {
	if host != "" {
		client.ServiceInfo.Host = host
	}
}

func (client *Client) SetScheme(scheme string) {
	if scheme != "" {
		client.ServiceInfo.Scheme = scheme
	}
}

// SetCredential 设置Credentials
func (client *Client) SetCredential(c Credentials) {
	if c.AccessKeyID != "" {
		client.ServiceInfo.Credentials.AccessKeyID = c.AccessKeyID
	}

	if c.SecretAccessKey != "" {
		client.ServiceInfo.Credentials.SecretAccessKey = c.SecretAccessKey
	}

	if c.Region != "" {
		client.ServiceInfo.Credentials.Region = c.Region
	}

	if c.SessionToken != "" {
		client.ServiceInfo.Credentials.SessionToken = c.SessionToken
	}
}

func (client *Client) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		client.ServiceInfo.Timeout = timeout
	}
}

// GetSignUrl 获取签名字符串
func (client *Client) GetSignUrl(api string, query url.Values) (string, error) {
	apiInfo := client.ApiInfoList[api]

	if apiInfo == nil {
		return "", errors.New("相关api不存在")
	}

	query = mergeQuery(query, apiInfo.Query)

	u := url.URL{
		Scheme:   client.ServiceInfo.Scheme,
		Host:     client.ServiceInfo.Host,
		Path:     apiInfo.Path,
		RawQuery: query.Encode(),
	}
	req, err := http.NewRequest(strings.ToUpper(apiInfo.Method), u.String(), nil)

	if err != nil {
		return "", errors.New("构建request失败")
	}

	return client.ServiceInfo.Credentials.SignUrl(req), nil
}

// SignSts2 生成sts信息
func (client *Client) SignSts2(inlinePolicy *Policy, expire time.Duration) (*SecurityToken2, error) {
	var err error
	sts := new(SecurityToken2)
	if sts.AccessKeyID, sts.SecretAccessKey, err = createTempAKSK(); err != nil {
		return nil, err
	}

	if expire < time.Minute {
		expire = time.Minute
	}

	now := time.Now()
	expireTime := now.Add(expire)
	sts.CurrentTime = now.Format(time.RFC3339)
	sts.ExpiredTime = expireTime.Format(time.RFC3339)

	innerToken, err := createInnerToken(client.ServiceInfo.Credentials, sts, inlinePolicy, expireTime.Unix())
	if err != nil {
		return nil, err
	}

	b, _ := json.Marshal(innerToken)
	sts.SessionToken = "STS2" + base64.StdEncoding.EncodeToString(b)
	return sts, nil
}

// Query 发起Get的query请求
func (client *Client) Query(api string, query url.Values) ([]byte, int, error) {
	return client.requestWithContentType(api, query, "", "")
}

// Json 发起Json的post请求
func (client *Client) Json(api string, query url.Values, body string) ([]byte, int, error) {
	return client.requestWithContentType(api, query, body, "application/json")
}

// PostWithContentType 发起自定义 Content-Type 的 post 请求，Content-Type 不可以为空
func (client *Client) PostWithContentType(api string, query url.Values, body string, ct string) ([]byte, int, error) {
	return client.requestWithContentType(api, query, body, ct)
}

func (client *Client) requestWithContentType(api string, query url.Values, body string, ct string) ([]byte, int, error) {
	apiInfo := client.ApiInfoList[api]

	if apiInfo == nil {
		return []byte(""), 500, errors.New("相关api不存在")
	}
	timeout := getTimeout(client.ServiceInfo.Timeout, apiInfo.Timeout)
	header := mergeHeader(client.ServiceInfo.Header, apiInfo.Header)
	query = mergeQuery(query, apiInfo.Query)

	u := url.URL{
		Scheme:   client.ServiceInfo.Scheme,
		Host:     client.ServiceInfo.Host,
		Path:     apiInfo.Path,
		RawQuery: query.Encode(),
	}
	var requestBody io.Reader
	if body != "" {
		requestBody = strings.NewReader(body)
	}
	req, err := http.NewRequest(strings.ToUpper(apiInfo.Method), u.String(), requestBody)
	if err != nil {
		return []byte(""), 500, errors.New("构建request失败")
	}
	req.Header = header
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	return client.makeRequest(api, req, timeout)
}

// Post 发起Post请求
func (client *Client) Post(api string, query url.Values, form url.Values) ([]byte, int, error) {
	apiInfo := client.ApiInfoList[api]
	form = mergeQuery(form, apiInfo.Form)
	return client.requestWithContentType(api, query, form.Encode(), "application/x-www-form-urlencoded")
}

func (client *Client) makeRequest(api string, req *http.Request, timeout time.Duration) ([]byte, int, error) {
	req = client.ServiceInfo.Credentials.Sign(req)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Client.Do(req)
	if err != nil {
		return []byte(""), 500, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), resp.StatusCode, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return body, resp.StatusCode, fmt.Errorf("api %s http code %d body %s", api, resp.StatusCode, string(body))
	}

	return body, resp.StatusCode, nil
}
