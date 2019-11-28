package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	httpHeaderJSON        = "application/json"
	httpHeaderContentType = "Content-Type"
	httpHeaderAccept      = "Accept"
)

//APIClientConfig : Keep config options for APIClient
type APIClientConfig struct {
	Username string
	Password string
	CaFile   string
	CertFile string
	KeyFile  string
	Proxy    string
}

//APIClient provided the http client for trigger http requests
type APIClient struct {
	//http client
	client *http.Client

	//Configuration
	config APIClientConfig
}

//NewAPIClient is constructor of APIClient
func NewAPIClient(config APIClientConfig) (*APIClient, error) {
	//Load client cert
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, err
	}

	//Add ca
	caCert, err := ioutil.ReadFile(config.CaFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	//If proxy should be set
	if len(strings.TrimSpace(config.Proxy)) > 0 {
		if proxyURL, err := url.Parse(config.Proxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
	}

	return &APIClient{
		client: client,
		config: config,
	}, nil

}

//Get data
func (ac *APIClient) Get(url string) ([]byte, error) {
	if strings.TrimSpace(url) == "" {
		return nil, errors.New("empty url")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(httpHeaderAccept, httpHeaderJSON)
	req.SetBasicAuth(ac.config.Username, ac.config.Password)

	resp, err := ac.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//Post data
func (ac *APIClient) Post(url string, data []byte) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("Empty url")
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	req.Header.Set(httpHeaderContentType, httpHeaderJSON)
	req.SetBasicAuth(ac.config.Username, ac.config.Password)
	resp, err := ac.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		if err := getErrorMessage(resp); err != nil {
			return fmt.Errorf("%s:%s", resp.Status, err.Error())
		}

		return errors.New(resp.Status)
	}

	return nil
}

//Delete data
func (ac *APIClient) Delete(url string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("Empty url")
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set(httpHeaderAccept, httpHeaderJSON)
	req.SetBasicAuth(ac.config.Username, ac.config.Password)

	resp, err := ac.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if err := getErrorMessage(resp); err != nil {
			return fmt.Errorf("%s:%s", resp.Status, err.Error())
		}

		return errors.New(resp.Status)
	}

	return nil
}

//SwitchAccount : Switch account
func (ac *APIClient) SwitchAccount(username, password string) {
	if len(strings.TrimSpace(username)) == 0 ||
		len(strings.TrimSpace(password)) == 0 {
		return
	}

	ac.config.Username = username
	ac.config.Password = password
}

//Read error message from response body
func getErrorMessage(resp *http.Response) error {
	if resp == nil {
		return errors.New("nil response")
	}

	if resp.Body == nil || resp.ContentLength == 0 {
		//nothing to read
		return nil
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//abandon to read deatiled error message
		return nil
	}

	return fmt.Errorf("%s", data)
}
