// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package icon

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	// import the gif format
	_ "image/gif"
	// import the jpeg format
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"sync"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/nfnt/resize"
)

// const definitions
const (
	DigestOfIconImage   = "sha256:0048162a053eef4d4ce3fe7518615bef084403614f8bca43b40ae2e762e11e06"
	DigestOfIconChart   = "sha256:61cf3a178ff0f75bf08a25d96b75cf7355dc197749a9f128ed3ef34b0df05518"
	DigestOfIconCNAB    = "sha256:089bdda265c14d8686111402c8ad629e8177a1ceb7dcd0f7f39b6480f623b3bd"
	DigestOfIconDefault = "sha256:da834479c923584f4cbcdecc0dac61f32bef1d51e8aae598cf16bd154efab49f"
)

var (
	builtInIcons = map[string]string{
		DigestOfIconImage:   "./icons/image.png",
		DigestOfIconChart:   "./icons/chart.png",
		DigestOfIconCNAB:    "./icons/cnab.png",
		DigestOfIconDefault: "./icons/default.png",
	}
	// Ctl is a global icon controller instance
	Ctl = NewController()
)

// Icon model for artifact icon
type Icon struct {
	ContentType string
	Content     string // base64 encoded
}

// Controller defines the operations related with icon
type Controller interface {
	// Get the icon specified by digest
	Get(ctx context.Context, digest string) (icon *Icon, err error)
}

// NewController creates a new instance of the icon controller
func NewController() Controller {
	return &controller{
		artMgr: artifact.Mgr,
		regCli: registry.Cli,
		cache:  sync.Map{},
	}
}

type controller struct {
	artMgr artifact.Manager
	regCli registry.Client
	cache  sync.Map
}

func (c *controller) Get(ctx context.Context, digest string) (*Icon, error) {
	ic, exist := c.cache.Load(digest)
	if exist {
		return ic.(*Icon), nil
	}

	var (
		iconFile io.ReadCloser
		err      error
	)
	// for the fixed icons: image, helm chart, CNAB and unknown
	if path, exist := builtInIcons[digest]; exist {
		iconFile, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		defer iconFile.Close()
	} else {
		// read icon from blob
		artifacts, err := c.artMgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"Icon": digest,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(artifacts) == 0 {
			return nil, errors.New(nil).WithCode(errors.NotFoundCode).
				WithMessage("the icon %s not found", digest)
		}
		_, iconFile, err = c.regCli.PullBlob(artifacts[0].RepositoryName, digest)
		if err != nil {
			return nil, err
		}
		defer iconFile.Close()
	}

	img, _, err := image.Decode(iconFile)
	if err != nil {
		return nil, err
	}

	// resize the icon to 50x50
	img = resize.Thumbnail(50, 50, img, resize.NearestNeighbor)

	// encode the resized icon to png
	buf := &bytes.Buffer{}
	if err = png.Encode(buf, img); err != nil {
		return nil, err
	}

	icon := &Icon{
		ContentType: "image/png",
		Content:     base64.StdEncoding.EncodeToString(buf.Bytes()),
	}

	c.cache.Store(digest, icon)

	return icon, nil
}
