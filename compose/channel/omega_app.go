package channel

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/harbor/compose/command"
	"github.com/vmware/harbor/compose/compose"
	. "github.com/vmware/harbor/compose/util"
)

const AUTH_HEADER = "Authorization"

const FORCE_IMAGE_PULL = false

const (
	OMEGA_APP_CREATE_API = "/api/v3/clusters/%d/apps"
	OMEGA_APP_STATUS_API = "/api/v3/clusters/%d/apps/%d/status"
)

type OAEnvironment struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type OAVolume struct {
	HostPath      string `json:"hostPath"`
	ContainerPath string `json:"containerPath"`
	Mode          string `json:"mode"`
}
type OAPortMappings struct {
	AppPort  int    `json:"appPort"`
	Protocol int    `json:"protocol"`
	IsUri    int    `json:"isUri"`
	Type     int    `json:"type"`
	MapPort  int    `json:"mapPort"`
	Uri      string `json:"uri"`
}
type OALabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AppCreationRequest struct {
	Name         string            `json:"name"`
	Instances    int32             `json:"instances"`
	ImageName    string            `json:"imageName"`
	ImageVersion string            `json:"imageVersion"`
	Network      string            `json:"network"`
	Cpus         float32           `json:"cpus"`
	Mem          float32           `json:"mem""`
	Cmd          string            `json:"cmd"`
	ForceImage   bool              `json:"forceImage"`
	Envs         []*OAEnvironment  `json:"envs"`
	Labels       []*OALabel        `json:"labels"`
	Volumes      []*OAVolume       `json:"volumes"`
	PortMappings []*OAPortMappings `json:"portMappings"`
	Constraints  []string          // todo
	LogPaths     []string          `json:"logPaths"`
	Parameters   []string          `json:"parameters"`
}

type AppCreationResponse struct {
	Code int                      `json:"code"`
	Data *AppCreationResponseData `json:"data"`
}

type AppCreationResponseData struct {
	Id        int    `json:"id"`
	Name      string `json:"string"`
	Instances string `json:"instances"`
}

type OmegaClient struct {
	HttpClient    *http.Client
	ChannelConfig ChannelHttpConfig
}

type OmegaAppOutput struct {
	Client     *OmegaClient
	SryCompose *compose.SryCompose
}

func NewOmegaOutput(auth ChannelHttpConfig) *OmegaAppOutput {
	httpClient := &http.Client{}
	if strings.HasPrefix(auth.AppApiUrl, "https") {
		httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	}

	client := &OmegaClient{
		HttpClient:    httpClient,
		ChannelConfig: auth,
	}
	return &OmegaAppOutput{Client: client}
}

func (output *OmegaAppOutput) Run(sry_compose *compose.SryCompose, cmd command.Command) error {
	if cmd == command.CREATE_APP {
		output.Create(sry_compose, cmd)
	} else if cmd == command.STOP_APP {
	}
	return nil
}

func (output *OmegaAppOutput) Create(sry_compose *compose.SryCompose, cmd command.Command) error {
	for _, app := range sry_compose.Applications {
		log.Println("creating application ", app.Name)

		imageStruct, err := ParseImage(app.Image)
		if err != nil {
			return err
		}

		request := &AppCreationRequest{
			Name:         app.Name,
			Instances:    app.Instances,
			ImageName:    imageStruct.ImageName(),
			ImageVersion: imageStruct.Version,
			Network:      app.Net,
			Cpus:         app.Cpu,
			Mem:          app.Mem,
			Cmd:          app.FormatedCommand(),
			ForceImage:   FORCE_IMAGE_PULL,
			LogPaths:     app.LogPaths,
			Parameters:   []string{},
			PortMappings: []*OAPortMappings{},
			Envs:         []*OAEnvironment{},
			Constraints:  []string{},
		}

		// for omega app, AppName(user entered from ui) have higher priority comparing to app.Name
		if len(app.AppName) > 0 {
			request.Name = app.AppName
		}

		// for omega app, ImageVersion (user entered from ui) have higher priority comparing to version guessed
		if len(app.ImageVersion) > 0 {
			request.ImageVersion = app.ImageVersion
		}

		for _, v := range app.Environment {
			env := &OAEnvironment{
				Key:   v.Key,
				Value: v.Value,
			}
			request.Envs = append(request.Envs, env)
		}

		for _, v := range app.Volumes {
			volume := &OAVolume{
				HostPath:      v.Host,
				ContainerPath: v.Container,
				Mode:          strings.ToUpper(v.Permission),
			}
			request.Volumes = append(request.Volumes, volume)
		}

		for _, v := range app.Labels {
			label := &OALabel{
				Key:   v.Key,
				Value: v.Value,
			}
			request.Labels = append(request.Labels, label)
		}

		for _, v := range app.Ports {
			portMap := &OAPortMappings{
				AppPort:  v.ContainerPort,
				Protocol: _sry_protocol(v.Protocol),
				IsUri:    2,
				Type:     1,
				MapPort:  v.HostPort,
				Uri:      "",
			}
			request.PortMappings = append(request.PortMappings, portMap)
		}

		requestJson, _ := json.Marshal(request)
		resp, err := output.post(fmt.Sprintf(OMEGA_APP_CREATE_API, app.ClusterId), string(requestJson))
		if err != nil {
			log.Println(err.Error())
			return err
		}
		defer resp.Body.Close()

		var appCreationResponse AppCreationResponse
		err = json.NewDecoder(resp.Body).Decode(&appCreationResponse)
		if err != nil {
			return err
		}
		if appCreationResponse.Code != 0 {
			return errors.New("")
		}
	}
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

func (output *OmegaAppOutput) get(path string, values url.Values) (*http.Response, error) {
	log.Println("GET: " + path)
	log.Println("params payload: " + values.Encode())

	req, _ := http.NewRequest("GET", output.expandPath(path, values), nil)
	output._auth(req)
	resp, err := output.Client.HttpClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != 200 {
		return resp, errors.New("POST:" + path + "  GOT:" + resp.Status)
	}

	return resp, nil
}

func (output *OmegaAppOutput) post(path string, json string) (*http.Response, error) {
	log.Println("POST: " + path)
	log.Println("JSON payload: " + json)

	req, err := http.NewRequest("POST", output.expandPath(path, url.Values{}), strings.NewReader(json))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	output._auth(req)
	resp, err := output.Client.HttpClient.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != 200 {
		log.Println("error", err)
		return resp, errors.New("POST:" + path + "  GOT:" + resp.Status)
	}

	return resp, nil
}

func (output *OmegaAppOutput) expandPath(path string, values url.Values) string {
	host := output.Client.ChannelConfig.AppApiUrl
	if strings.HasSuffix(host, "/") {
		host = strings.TrimRight(host, "/")
	}

	if strings.HasPrefix(path, "/") {
		path = strings.TrimLeft(path, "/")
	}

	query := values.Encode()

	return host + "/" + path + "?" + query
}

func (output *OmegaAppOutput) _auth(req *http.Request) {
	if output.Client.ChannelConfig.Type == "token" {
		req.Header.Set(AUTH_HEADER, output.Client.ChannelConfig.Token)
	} else if output.Client.ChannelConfig.Type == "http_basic" {
		req.SetBasicAuth(output.Client.ChannelConfig.Principle, output.Client.ChannelConfig.Password)
	}
}

func _sry_protocol(protocol string) int {
	if protocol == "http" {
		return 2
	} else {
		return 1
	}
}
