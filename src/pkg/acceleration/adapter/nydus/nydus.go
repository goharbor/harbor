package nydus

import (
	"bytes"
	"encoding/json"
	"fmt"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/acceleration/adapter"
	"github.com/goharbor/harbor/src/pkg/acceleration/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"net/http"
	"time"
)

func init() {
	if err := adp.RegisterFactory(model.AccelerationTypeNydus, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.AccelerationTypeNydus, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.AccelerationTypeNydus)
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.AccelerationService) (adp.Adapter, error) {
	return &adapter{
		url: r.URL,
	}, nil
}

var (
	_ adp.Adapter = (*adapter)(nil)
)

type adapter struct {
	url string
}

func (a *adapter) Convert(art *artifact.Artifact, tag string) error {
	hc := &http.Client{}
	addr := fmt.Sprintf("%s/api/v1/conversions", a.url)
	payload, err := a.getPayload(art, tag)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, addr, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("nydus job(target: %s) response code is %d", addr, resp.StatusCode)
	}

	return nil
}

func (a *adapter) HealthCheck() (string, error) {
	return model.Healthy, nil
}

func (a *adapter) getPayload(art *artifact.Artifact, tag string) ([]byte, error) {
	url, err := BuildImageResourceURL(art.RepositoryName, tag)
	if err != nil {
		return []byte{}, err
	}
	payload := &notifyModel.Payload{
		Type:     event2.TopicPushArtifact,
		Operator: "admin",
		OccurAt:  time.Now().Unix(),
		EventData: &notifyModel.EventData{
			Resources: []*notifyModel.Resource{
				{
					Digest:      art.Digest,
					Tag:         tag,
					ResourceURL: url,
				},
			},
		},
	}

	return json.Marshal(payload)
}

// BuildImageResourceURL ...
func BuildImageResourceURL(repoName, reference string) (string, error) {
	extURL, err := config.ExtURL()
	if err != nil {
		return "", fmt.Errorf("get external endpoint failed: %v", err)
	}

	if distribution.IsDigest(reference) {
		return fmt.Sprintf("%s/%s@%s", extURL, repoName, reference), nil
	}

	return fmt.Sprintf("%s/%s:%s", extURL, repoName, reference), nil
}
