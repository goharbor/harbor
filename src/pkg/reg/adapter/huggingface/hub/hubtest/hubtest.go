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

// Package hubtest provides an in-memory fake of the Hugging Face Hub API for tests.
package hubtest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
)

// File is a file of a fake commit.
type File struct {
	Content []byte
	LFS     bool
}

// Commit is one fake commit.
type Commit struct {
	Files        map[string]File
	LastModified time.Time
	License      any
}

// Model is one fake model.
type Model struct {
	// ID is the canonical model ID.
	ID       string
	Commits  map[string]*Commit
	Branches map[string]string
	Tags     map[string]string
	// Gated models answer 401 to requests without the token.
	Gated bool
}

// Hub is a fake Hub.
type Hub struct {
	*httptest.Server
	Token string
	// PageSize is the page size of the model listing.
	PageSize int

	mu        sync.Mutex
	models    map[string]*Model
	calls     map[string]int
	rateLimit map[string]int
	failures  map[string]int
}

// New starts a fake Hub.
func New(models ...*Model) *Hub {
	h := &Hub{
		PageSize:  2,
		models:    map[string]*Model{},
		calls:     map[string]int{},
		rateLimit: map[string]int{},
		failures:  map[string]int{},
	}
	for _, m := range models {
		h.models[strings.ToLower(m.ID)] = m
	}
	h.Server = httptest.NewServer(http.HandlerFunc(h.serve))
	return h
}

// Calls returns how often an endpoint kind was hit: revision, refs, list, resolve, whoami, account.
func (h *Hub) Calls(kind string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.calls[kind]
}

// ResetCalls clears the call counters.
func (h *Hub) ResetCalls() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.calls = map[string]int{}
}

// RateLimit makes the next n requests of an endpoint kind answer 429 with Retry-After: 0.
func (h *Hub) RateLimit(kind string, n int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rateLimit[kind] = n
}

// Fail makes the next n requests of an endpoint kind answer 500.
func (h *Hub) Fail(kind string, n int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failures[kind] = n
}

// SetTag moves a tag of a model.
func (h *Hub) SetTag(modelID, tag, commit string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.models[strings.ToLower(modelID)].Tags[tag] = commit
}

// SetBranch moves a branch of a model.
func (h *Hub) SetBranch(modelID, branch, commit string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.models[strings.ToLower(modelID)].Branches[branch] = commit
}

// count records a call and reports whether it was answered with an injected error.
func (h *Hub) count(w http.ResponseWriter, kind string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.calls[kind]++
	if h.failures[kind] > 0 {
		h.failures[kind]--
		hubError(w, http.StatusInternalServerError, "", "internal error")
		return true
	}
	if h.rateLimit[kind] > 0 {
		h.rateLimit[kind]--
		tooMany(w)
		return true
	}
	return false
}

func (h *Hub) model(id string) *Model {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.models[strings.ToLower(id)]
}

func (h *Hub) serve(w http.ResponseWriter, r *http.Request) {
	p := r.URL.EscapedPath()
	switch {
	case p == "/api/whoami-v2":
		if h.count(w, "whoami") {
			return
		}
		if h.Token == "" || r.Header.Get("Authorization") != "Bearer "+h.Token {
			hubError(w, http.StatusUnauthorized, "", "Invalid username or password.")
			return
		}
		writeJSON(w, map[string]string{"name": "tester"})
	case p == "/api/models":
		if h.count(w, "list") {
			return
		}
		h.list(w, r)
	case strings.HasPrefix(p, "/api/organizations/") || strings.HasPrefix(p, "/api/users/"):
		if h.count(w, "account") {
			return
		}
		h.account(w, p)
	case strings.HasPrefix(p, "/api/models/"):
		h.modelAPI(w, r, strings.TrimPrefix(p, "/api/models/"))
	default:
		h.resolve(w, r, strings.TrimPrefix(p, "/"))
	}
}

func (h *Hub) modelAPI(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.SplitN(rest, "/", 4)
	if len(parts) < 3 {
		hubError(w, http.StatusNotFound, "", "not found")
		return
	}
	id := parts[0] + "/" + parts[1]
	kind := parts[2]
	if h.count(w, kind) {
		return
	}
	m, ok := h.authorized(w, r, id)
	if !ok {
		return
	}
	if m.ID != id {
		redirect(w, r, "/api/models/"+m.ID+"/"+strings.Join(parts[2:], "/"))
		return
	}
	switch {
	case kind == "refs" && len(parts) == 3:
		writeJSON(w, map[string]any{"branches": refs(m.Branches), "tags": refs(m.Tags), "converts": []any{}})
	case kind == "revision" && len(parts) == 4:
		rev, err := url.PathUnescape(parts[3])
		if err != nil {
			hubError(w, http.StatusBadRequest, "", err.Error())
			return
		}
		h.revision(w, m, rev)
	default:
		hubError(w, http.StatusNotFound, "", "not found")
	}
}

func (h *Hub) revision(w http.ResponseWriter, m *Model, rev string) {
	h.mu.Lock()
	commit := rev
	if c, ok := m.Branches[rev]; ok {
		commit = c
	} else if c, ok := m.Tags[rev]; ok {
		commit = c
	}
	c, ok := m.Commits[commit]
	h.mu.Unlock()
	if !ok {
		hubError(w, http.StatusNotFound, "RevisionNotFound", "Invalid rev id: "+rev)
		return
	}
	paths := make([]string, 0, len(c.Files))
	for p := range c.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	siblings := make([]map[string]any, 0, len(paths))
	for _, p := range paths {
		f := c.Files[p]
		sib := map[string]any{"rfilename": p, "size": len(f.Content)}
		if f.LFS {
			sum := sha256.Sum256(f.Content)
			sib["blobId"] = hub.GitBlobID([]byte("pointer " + p))
			sib["lfs"] = map[string]any{"sha256": hex.EncodeToString(sum[:]), "size": len(f.Content), "pointerSize": 135}
		} else {
			sib["blobId"] = hub.GitBlobID(f.Content)
		}
		siblings = append(siblings, sib)
	}
	resp := map[string]any{
		"id":           m.ID,
		"modelId":      m.ID,
		"sha":          commit,
		"lastModified": c.LastModified.UTC().Format("2006-01-02T15:04:05.000Z"),
		"downloads":    time.Now().UnixNano(),
		"siblings":     siblings,
	}
	if c.License != nil {
		resp["cardData"] = map[string]any{"license": c.License}
	}
	writeJSON(w, resp)
}

func (h *Hub) resolve(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.SplitN(rest, "/", 5)
	if len(parts) != 5 || parts[2] != "resolve" {
		hubError(w, http.StatusNotFound, "", "not found")
		return
	}
	if h.count(w, "resolve") {
		return
	}
	id := parts[0] + "/" + parts[1]
	m, ok := h.authorized(w, r, id)
	if !ok {
		return
	}
	filePath, err := url.PathUnescape(parts[4])
	if err != nil {
		hubError(w, http.StatusBadRequest, "", err.Error())
		return
	}
	h.mu.Lock()
	c, ok := m.Commits[parts[3]]
	var f File
	if ok {
		f, ok = c.Files[filePath]
	}
	h.mu.Unlock()
	if !ok {
		hubError(w, http.StatusNotFound, "EntryNotFound", "Entry not found")
		return
	}
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(f.Content))
}

func (h *Hub) list(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	var ids []string
	h.mu.Lock()
	for _, m := range h.models {
		if author == "" || strings.SplitN(m.ID, "/", 2)[0] == author {
			ids = append(ids, m.ID)
		}
	}
	h.mu.Unlock()
	sort.Strings(ids)
	start, _ := strconv.Atoi(r.URL.Query().Get("cursor"))
	end := min(start+h.PageSize, len(ids))
	if start > len(ids) {
		start = end
	}
	if end < len(ids) {
		q := r.URL.Query()
		q.Set("cursor", strconv.Itoa(end))
		w.Header().Set("Link", `<`+h.URL+`/api/models?`+q.Encode()+`>; rel="next"`)
	}
	page := []map[string]string{}
	for _, id := range ids[start:end] {
		page = append(page, map[string]string{"id": id, "modelId": id})
	}
	writeJSON(w, page)
}

func (h *Hub) account(w http.ResponseWriter, p string) {
	name := strings.Split(strings.TrimPrefix(strings.TrimPrefix(p, "/api/organizations/"), "/api/users/"), "/")[0]
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, m := range h.models {
		author := strings.SplitN(m.ID, "/", 2)[0]
		if strings.EqualFold(author, name) {
			writeJSON(w, map[string]string{"name": author})
			return
		}
	}
	hubError(w, http.StatusNotFound, "", "This user does not exist")
}

func (h *Hub) authorized(w http.ResponseWriter, r *http.Request, id string) (*Model, bool) {
	m := h.model(id)
	if m == nil {
		// The real Hub answers 401, not 404, for unknown repositories.
		hubError(w, http.StatusUnauthorized, "", "Invalid username or password.")
		return nil, false
	}
	if m.Gated && (h.Token == "" || r.Header.Get("Authorization") != "Bearer "+h.Token) {
		hubError(w, http.StatusUnauthorized, "GatedRepo", "Access to model "+m.ID+" is restricted.")
		return nil, false
	}
	return m, true
}

func refs(names map[string]string) []map[string]string {
	out := []map[string]string{}
	keys := make([]string, 0, len(names))
	for k := range names {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, map[string]string{"name": k, "ref": "refs/heads/" + k, "targetCommit": names[k]})
	}
	return out
}

func redirect(w http.ResponseWriter, r *http.Request, target string) {
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func tooMany(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "0")
	hubError(w, http.StatusTooManyRequests, "", "rate limited")
}

func hubError(w http.ResponseWriter, status int, code, msg string) {
	if code != "" {
		w.Header().Set("X-Error-Code", code)
	}
	w.Header().Set("X-Error-Message", msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
