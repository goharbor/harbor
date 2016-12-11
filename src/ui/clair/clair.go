package clair

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/service/token"

	clairapiv1 "github.com/coreos/clair/api/v1"
	clairtypes "github.com/coreos/clair/utils/types"
)

var SeverityLevels = []clairtypes.Priority{clairtypes.Low, clairtypes.Medium, clairtypes.High, clairtypes.Critical}

const (
	tokenService        = "token-service"
	tokenScopeTemplate  = "repository:%s:pull"
	emptyLayerBlobSum   = "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
	getLayerFeaturesURI = "%s/layers/%s?vulnerabilities"
)

var clairClient *Clair = nil

// Clair represents the Clair server
type Clair struct {
	url string
}

type layer struct {
	Name       string
	Path       string
	ParentName string
	Format     string
	Headers    headers
}

type headers struct {
	Authorization string
}

type errorResponse struct {
	Message string
}

type layerEnvelope struct {
	Layer *layer `json:"Layer,omitempty"`
}

func init() {
	clairUrl := config.ClairURL()

	clairClient = &Clair{clairUrl}
}

func AnalyseImage(user string, manifest distribution.Manifest, repoName string) {
	log.Infof("Start to analyse the security for image")
	ui_url := config.ExternalUiURL()
	registryUrl := ui_url + "/v2"

	clairClient.Analyse(user, manifest, repoName, registryUrl)
}

func GetSecurity(manifest distribution.Manifest) (clairapiv1.Layer, error) {
	// Get the security according the top layer in Manifest schema2
	topLayer := manifest.References()[0].Digest.String()

	// TODO (robin) Classify and sort the vulnerabilities of the image before return
	return clairClient.getLayer(topLayer)

}

// Analyse Sends each layer from Docker image to Clair
func (c *Clair) Analyse(username string, manifest distribution.Manifest, repoName, registryUrl string) {
	// Make the token
	scope := fmt.Sprintf(tokenScopeTemplate, repoName)
	access := token.GetResourceActions([]string{scope})
	token, _, _, err := token.MakeToken(username, tokenService, access)
	if err != nil {
		log.Errorf("Failed to make token, error: %v", err)
		return
	}

	refs := manifest.References()
	for i := len(refs) - 1; i >= 0; i-- {
		if refs[i].Digest.String() == emptyLayerBlobSum {
			// Ignore the empty layer
			log.Infof("Ignore the empty layer.")
			continue
		}

		layer := newLayer(refs, i, repoName, registryUrl, token)
		err := c.pushLayer(layer)
		if err != nil {
			log.Errorf("Push layer %d failed: %s", i, err.Error())
			continue
		}
	}
}

func (c *Clair) pushLayer(layer *layer) error {
	envelope := layerEnvelope{Layer: layer}
	reqBody, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("can't serialze push request: %s", err)
	}
	url := fmt.Sprintf("%s/layers", c.url)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("Can't create a push request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: time.Minute}).Do(request)
	if err != nil {
		return fmt.Errorf("Can't push layer to Clair: %s", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Can't read clair response : %s", err)
	}
	if response.StatusCode != http.StatusCreated {
		var errResp errorResponse
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return fmt.Errorf("Can't even read an error message: %s", err)
		}
		return fmt.Errorf("Push error %d: %s", response.StatusCode, string(body))
	}
	return nil
}

// GetLayer Gets the vulnerabilities according the layer id
func (c *Clair) getLayer(layerId string) (clairapiv1.Layer, error) {
	emptyLayer := clairapiv1.Layer{}

	response, err := http.Get(fmt.Sprintf(getLayerFeaturesURI, c.url, layerId))
	if err != nil {
		return emptyLayer, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		err := fmt.Errorf("Got response %d with message %s", response.StatusCode, string(body))
		return emptyLayer, err
	}

	var apiResponse clairapiv1.LayerEnvelope
	if err = json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		return emptyLayer, err
	} else if apiResponse.Error != nil {
		return emptyLayer, errors.New(apiResponse.Error.Message)
	}

	return *apiResponse.Layer, nil
}

func newLayer(refs []distribution.Descriptor, index int, repoName, registryUrl, token string) *layer {
	var parentName string
	maxIndex := len(refs) - 1
	for i := index; i < maxIndex; i++ {
		if refs[i+1].Digest.String() != emptyLayerBlobSum {
			parentName = refs[i+1].Digest.String()
			break
		}
	}

	tmpLayer := &layer{
		Name:       refs[index].Digest.String(),
		Path:       strings.Join([]string{registryUrl, repoName, "blobs", refs[index].Digest.String()}, "/"),
		ParentName: parentName,
		Format:     "Docker",
		Headers:    headers{fmt.Sprintf("Bearer %s", token)},
	}

	return tmpLayer
}
