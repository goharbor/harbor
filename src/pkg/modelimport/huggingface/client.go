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

package huggingface

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/goharbor/harbor/src/lib/retry"
)

const defaultEndpoint = "https://huggingface.co"

// Client reads model metadata and files from Hugging Face Hub.
type Client interface {
	Snapshot(ctx context.Context, repoID, revision, subpath, token string) (*Snapshot, error)
	OpenFile(ctx context.Context, repoID, revision, filename, token string) (io.ReadCloser, error)
}

// File describes a Hugging Face repository file.
type File struct {
	Path string
	Size int64
}

// Snapshot contains the metadata needed to convert a Hugging Face model.
type Snapshot struct {
	RepoID     string
	Revision   string
	CommitSHA  string
	Files      []File
	Readme     []byte
	CardData   map[string]any
	SourceURL  string
	SourceType string
}

// NewClient returns the default Hugging Face client.
func NewClient() Client {
	return &client{
		endpoint: defaultEndpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		fileClient: &http.Client{},
	}
}

type client struct {
	endpoint   string
	httpClient *http.Client
	fileClient *http.Client
}

type modelInfoResponse struct {
	SHA      string `json:"sha"`
	Siblings []struct {
		RFilename string `json:"rfilename"`
		Size      int64  `json:"size"`
	} `json:"siblings"`
	CardData map[string]any `json:"cardData"`
}

func (c *client) Snapshot(ctx context.Context, repoID, revision, subpath, token string) (*Snapshot, error) {
	if revision == "" {
		revision = "main"
	}
	info, err := c.modelInfo(ctx, repoID, revision, token)
	if err != nil {
		return nil, err
	}
	subpath = strings.Trim(strings.TrimSpace(subpath), "/")
	snapshot := &Snapshot{
		RepoID:     repoID,
		Revision:   revision,
		CommitSHA:  info.SHA,
		CardData:   info.CardData,
		SourceURL:  strings.TrimRight(c.endpoint, "/") + "/" + repoID,
		SourceType: "huggingface",
	}
	for _, sibling := range info.Siblings {
		filename := path.Clean(sibling.RFilename)
		if filename == "." || strings.HasPrefix(filename, "../") {
			continue
		}
		if subpath != "" {
			if filename != subpath && !strings.HasPrefix(filename, subpath+"/") {
				continue
			}
			filename = strings.TrimPrefix(filename, subpath+"/")
		}
		if strings.EqualFold(filename, "README.md") {
			if snapshot.Readme, err = c.readSmallFile(ctx, repoID, revision, sibling.RFilename, token, 4*1024*1024); err != nil {
				return nil, err
			}
			if len(snapshot.CardData) == 0 {
				snapshot.CardData = parseFrontMatter(snapshot.Readme)
			}
		}
		snapshot.Files = append(snapshot.Files, File{
			Path: filename,
			Size: sibling.Size,
		})
	}
	return snapshot, nil
}

func (c *client) OpenFile(ctx context.Context, repoID, revision, filename, token string) (io.ReadCloser, error) {
	var body io.ReadCloser
	err := retry.Retry(func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resolveURL(repoID, revision, filename), nil)
		if err != nil {
			return retry.Abort(err)
		}
		authorize(req, token)
		resp, err := c.streamClient().Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return retry.Abort(err)
			}
			if isTransientNetErr(err) {
				return err
			}
			return retry.Abort(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			defer resp.Body.Close()
			err = fmt.Errorf("download %s failed: %s", filename, resp.Status)
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
				return err
			}
			return retry.Abort(err)
		}
		body = resp.Body
		return nil
	}, retry.InitialInterval(500*time.Millisecond), retry.MaxInterval(5*time.Second), retry.Timeout(45*time.Second))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *client) modelInfo(ctx context.Context, repoID, revision, token string) (*modelInfoResponse, error) {
	u := strings.TrimRight(c.endpoint, "/") + "/api/models/" + repoID + "/revision/" + url.PathEscape(revision)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	authorize(req, token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Hugging Face model lookup failed: %s", resp.Status)
	}
	var info modelInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if info.SHA == "" {
		info.SHA = revision
	}
	return &info, nil
}

func (c *client) readSmallFile(ctx context.Context, repoID, revision, filename, token string, limit int64) ([]byte, error) {
	reader, err := c.OpenFile(ctx, repoID, revision, filename, token)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s exceeds %d byte limit", filename, limit)
	}
	return data, nil
}

func (c *client) resolveURL(repoID, revision, filename string) string {
	parts := strings.Split(filename, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.TrimRight(c.endpoint, "/") + "/" + repoID + "/resolve/" + url.PathEscape(revision) + "/" + strings.Join(parts, "/")
}

func (c *client) streamClient() *http.Client {
	if c.fileClient != nil {
		return c.fileClient
	}
	return c.httpClient
}

func authorize(req *http.Request, token string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func isTransientNetErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"unexpected eof",
		"connection reset",
		"broken pipe",
		"connection refused",
		"tls handshake timeout",
		"i/o timeout",
		"temporary failure",
		"http2: server sent goaway",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func parseFrontMatter(readme []byte) map[string]any {
	readme = bytes.TrimPrefix(readme, []byte("\xef\xbb\xbf"))
	if !bytes.HasPrefix(readme, []byte("---\n")) && !bytes.HasPrefix(readme, []byte("---\r\n")) {
		return nil
	}
	parts := bytes.SplitN(readme, []byte("---"), 3)
	if len(parts) < 3 {
		return nil
	}
	out := map[string]any{}
	if err := yaml.Unmarshal(parts[1], &out); err != nil {
		return nil
	}
	return out
}
