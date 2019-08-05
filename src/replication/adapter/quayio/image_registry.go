package quayio

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	registry_pkg "github.com/goharbor/harbor/src/common/utils/registry"
)

// PullManifest get manifest details from Quay.io
func (a *adapter) PullManifest(repository, reference string, accepttedMediaTypes []string) (distribution.Manifest, string, error) {
	req, err := http.NewRequest(http.MethodGet, buildManifestURL(repository, reference), nil)
	if err != nil {
		return nil, reference, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, reference, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, reference, fmt.Errorf("[%d][%s]", resp.StatusCode, string(body))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, reference, err
	}

	dm, _, err := registry_pkg.UnMarshal(schema1.MediaTypeSignedManifest, data)
	if err != nil {
		return dm, reference, err
	}
	return dm, reference, nil
}

// type layer struct {
// 	Comment         string   `json:"comment"`
// 	Index           int      `json:"index"`
// 	CompressedSize  int      `json:"compressed_size"`
// 	Command         []string `json:"command"`
// 	Urls            string   `json:"urls"`
// 	Author          string   `json:"author"`
// 	IsRemote        bool     `json:"is_remote"`
// 	CreatedDatetime string   `json:"created_datetime"`
// 	BlobDigest      string   `json:"blob_digest"`
// }

// type image struct {
// 	Comment   string   `json:"comment"`
// 	Ancestors string   `json:"ancestors"`
// 	Created   string   `json:"created"`
// 	Uploading bool     `json:"uploading"`
// 	SortIndex int      `json:"sort_index"`
// 	Command   []string `json:"command"`
// 	History   []image  `json:"history"`
// 	ID        string   `json:"id"`
// 	Size      string   `json:"size"`
// }

// type manifest struct {
// 	Layers         []layer `json:"layers"`
// 	Image          image   `json:"image"`
// 	ManifestData   string  `json:"manifest_data"`
// 	Digest         string  `json:"digest"`
// 	IsManifestList bool    `json:"is_manifest_list"`
// }
