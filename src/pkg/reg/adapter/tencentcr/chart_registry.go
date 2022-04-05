package tencentcr

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	chartListURL    = "%s/api/chartrepo/%s/charts"
	chartVersionURL = "%s/api/chartrepo/%s/charts/%s"
	chartInfoURL    = "%s/api/chartrepo/%s/charts/%s/%s"
)

type tcrChart struct {
	APIVersion string   `json:"apiVersion"`
	Digest     string   `json:"digest"`
	Name       string   `json:"name"`
	URLs       []string `json:"urls"`
	Version    string   `json:"version"`
}

type tcrChartVersionDetail struct {
	Metadata *tcrChartVersionMetadata `json:"metadata"`
}
type tcrChartVersionMetadata struct {
	URLs []string `json:"urls"`
}

var _ adp.ChartRegistry = &adapter{}

func (a *adapter) FetchCharts(filters []*model.Filter) (resources []*model.Resource, err error) {
	log.Debugf("[tencent-tcr.FetchCharts]filters: %#v", filters)
	// 1. list namespaces via TCR Special API
	var nsPattern, _, _ = filterToPatterns(filters)
	var nms []string
	nms, err = a.listCandidateNamespaces(nsPattern)
	if err != nil {
		return
	}

	return a.fetchCharts(nms, filters)
}

func (a *adapter) fetchCharts(namespaces []string, filters []*model.Filter) (resources []*model.Resource, err error) {
	// 1. list repositories
	for _, ns := range namespaces {
		var url = fmt.Sprintf(chartListURL, a.registry.URL, ns)
		var repositories = []*model.Repository{}
		err = a.client.Get(url, &repositories)
		log.Debugf("[tencent-tcr.FetchCharts] url=%s, namespace=%s, repositories=%v, error=%v", url, ns, repositories, err)
		if err != nil {
			return
		}
		if len(repositories) == 0 {
			continue
		}
		for _, repository := range repositories {
			repository.Name = fmt.Sprintf("%s/%s", ns, repository.Name)
		}
		repositories, err = filter.DoFilterRepositories(repositories, filters)
		if err != nil {
			return
		}

		// 2. list versions
		for _, repository := range repositories {
			var name = strings.SplitN(repository.Name, "/", 2)[1]
			var url = fmt.Sprintf(chartVersionURL, a.registry.URL, ns, name)
			var charts = []*tcrChart{}
			err = a.client.Get(url, &charts)
			if err != nil {
				return nil, err
			}
			if len(charts) == 0 {
				continue
			}
			var artifacts []*model.Artifact
			for _, chart := range charts {
				artifacts = append(artifacts, &model.Artifact{
					Tags: []string{chart.Version},
				})
			}
			artifacts, err = filter.DoFilterArtifacts(artifacts, filters)
			if err != nil {
				return nil, err
			}
			if len(artifacts) == 0 {
				continue
			}

			for _, artifact := range artifacts {
				resources = append(resources, &model.Resource{
					Type:     model.ResourceTypeChart,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: repository.Name,
						},
						Artifacts: []*model.Artifact{artifact},
					},
				})
			}
		}

	}

	return
}

func (a *adapter) ChartExist(name, version string) (exist bool, err error) {
	log.Debugf("[tencent-tcr.ChartExist] name=%s version=%s", name, version)
	_, err = a.getChartInfo(name, version)
	// if not found, return not exist
	if httpErr, ok := err.(*commonhttp.Error); ok && httpErr.Code == http.StatusNotFound {
		return false, nil
	}
	if err != nil {
		return
	}
	exist = true

	return
}

func (a *adapter) getChartInfo(name, version string) (info *tcrChartVersionDetail, err error) {
	var namespace string
	var chart string
	namespace, chart, err = parseChartName(name)
	if err != nil {
		return
	}

	var url = fmt.Sprintf(chartInfoURL, a.registry.URL, namespace, chart, version)
	info = &tcrChartVersionDetail{}
	err = a.client.Get(url, info)
	if err != nil {
		return
	}
	return
}

func (a *adapter) DownloadChart(name, version, contentURL string) (rc io.ReadCloser, err error) {
	var info *tcrChartVersionDetail
	info, err = a.getChartInfo(name, version)
	if err != nil {
		return
	}
	if info.Metadata == nil || len(info.Metadata.URLs) == 0 || len(info.Metadata.URLs[0]) == 0 {
		return nil, fmt.Errorf("[tencent-tcr.DownloadChart.NO_DOWNLOAD_URL] chart=%s:%s", name, version)
	}

	var url = strings.ToLower(info.Metadata.URLs[0])
	// relative URL
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		var namespace string
		namespace, _, err = parseChartName(name)
		if err != nil {
			return
		}
		url = fmt.Sprintf("%s/chartrepo/%s/%s", a.registry.URL, namespace, url)
	}

	var req *http.Request
	var resp *http.Response
	var body []byte
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	resp, err = a.client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		err = errors.Errorf("[tencent-tcr.DownloadChart.failed] chart=%q, status=%d, body=%s", req.URL.String(), resp.StatusCode, string(body))

		return
	}

	return resp.Body, nil
}

func (a *adapter) UploadChart(name, version string, reader io.Reader) (err error) {
	var namespace string
	var chart string
	namespace, chart, err = parseChartName(name)
	if err != nil {
		return
	}

	// 1. write to form-data buffer
	var buf = &bytes.Buffer{}
	var writer = multipart.NewWriter(buf)
	var fw io.Writer
	fw, err = writer.CreateFormFile("chart", chart+".tgz")
	if err != nil {
		return
	}
	_, err = io.Copy(fw, reader)
	if err != nil {
		return
	}
	writer.Close()

	// 2. upload
	var url = fmt.Sprintf(chartListURL, a.registry.URL, namespace)
	var req *http.Request
	var resp *http.Response
	req, err = http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err = a.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// 3. parse response
	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > 299 {
		err = &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
		return
	}
	return
}

func (a *adapter) DeleteChart(name, version string) (err error) {
	var namespace string
	var chart string
	namespace, chart, err = parseChartName(name)
	if err != nil {
		return
	}

	var url = fmt.Sprintf(chartInfoURL, a.registry.URL, namespace, chart, version)

	return a.client.Delete(url)
}

func parseChartName(name string) (namespace, chart string, err error) {
	strs := strings.Split(name, "/")
	if len(strs) == 2 && len(strs[0]) > 0 && len(strs[1]) > 0 {
		return strs[0], strs[1], nil
	}
	return "", "", fmt.Errorf("[tencent-tcr.parseChartName.invalid_name] name=%s", name)
}
