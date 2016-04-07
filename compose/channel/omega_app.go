package channel

import (
	"errors"
	"github.com/vmware/harbor/compose/command"
	"github.com/vmware/harbor/compose/compose"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const AUTH_HEADER = "ChannelConfigorization"

type OmegaClient struct {
	HttpClient    *http.Client
	ChannelConfig ChannelHttpConfig
}

type OmegaAppOutput struct {
	Client     *OmegaClient
	SryCompose *compose.SryCompose
}

func NewOmegaOutput(auth ChannelHttpConfig) *OmegaAppOutput {
	client := &OmegaClient{
		HttpClient:    &http.Client{},
		ChannelConfig: auth,
	}
	return &OmegaAppOutput{Client: client}
}

func (output *OmegaAppOutput) Run(sry_compose *compose.SryCompose, cmd command.Command) error {
	return nil
}

func (output *OmegaAppOutput) Create() error {
	return nil
}

func (output *OmegaAppOutput) Stop() error {
	return nil
}

func (output *OmegaAppOutput) Scale() error {
	return nil
}

func (output *OmegaAppOutput) Restart() error {
	return nil
}

func (output *OmegaAppOutput) get(path string, values url.Values) error {
	log.Println("GET: " + path)
	log.Println("params payload: " + values.Encode())

	req, _ := http.NewRequest("GET", path, nil)
	output._auth(req)
	resp, err := output.Client.HttpClient.Do(req)

	if resp.StatusCode != 200 {
		return errors.New("POST:" + path + "  GOT:" + resp.Status)
	}

	if err != nil {
		return err
	}

	return nil
}

func (output *OmegaAppOutput) post(path string, json string) error {
	log.Println("POST: " + path)
	log.Println("JSON payload: " + json)

	req, _ := http.NewRequest("POST", path, strings.NewReader(json))
	output._auth(req)
	resp, err := output.Client.HttpClient.Do(req)

	if resp.StatusCode != 200 {
		return errors.New("POST:" + path + "  GOT:" + resp.Status)
	}

	if err != nil {
		return err
	}

	return nil
}

func (output *OmegaAppOutput) _auth(req *http.Request) {
	if output.Client.ChannelConfig.Type == "token" {
		req.Header.Set(AUTH_HEADER, output.Client.ChannelConfig.Token)
	} else if output.Client.ChannelConfig.Type == "http_basic" {
		req.SetBasicAuth(output.Client.ChannelConfig.Principle, output.Client.ChannelConfig.Password)
	} else {
	}
}
