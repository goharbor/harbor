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

package registry

import (
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/server/router"
	"github.com/opencontainers/go-digest"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	pullCounterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_pull_total",
			Help: "Number of docker pull operations",
		},
		[]string{"repository", "reference"},
	)

	pullCounterSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_pull_success",
			Help: "Number of successful docker pull operations",
		},
		[]string{"repository", "reference"},
	)

	pullCounterFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_pull_failed",
			Help: "Number of failed docker pull operations",
		},
		[]string{"repository", "reference"},
	)

	pushCounterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_push_total",
			Help: "Number of docker push operations",
		},
		[]string{"repository", "reference"},
	)

	pushCounterSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_push_success",
			Help: "Number of successful docker push operations",
		},
		[]string{"repository", "reference"},
	)

	pushCounterFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_push_failed",
			Help: "Number of failed docker push operations",
		},
		[]string{"repository", "reference"},
	)

	deleteCounterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_delete_total",
			Help: "Number of manifest delete operations",
		},
		[]string{"repository", "reference"},
	)

	deleteCounterSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_delete_success",
			Help: "Number of successful manifest delete operations",
		},
		[]string{"repository", "reference"},
	)

	deleteCounterFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "harbor_delete_failed",
			Help: "Number of failed manifest delete operations",
		},
		[]string{"repository", "reference"},
	)
)

func init() {
	prometheus.MustRegister(
		pullCounterTotal,
		pullCounterSuccess,
		pullCounterFailed,
		pushCounterTotal,
		pushCounterSuccess,
		pushCounterFailed,
		deleteCounterTotal,
		deleteCounterSuccess,
		deleteCounterFailed,
	)
}

// make sure the artifact exist before proxying the request to the backend registry
func getManifest(w http.ResponseWriter, req *http.Request) {
	repository := router.Param(req.Context(), ":splat")
	reference := router.Param(req.Context(), ":reference")
	pullCounterTotal.WithLabelValues(repository, reference).Inc()
	art, err := artifact.Ctl.GetByReference(req.Context(), repository, reference, nil)
	if err != nil {
		lib_http.SendError(w, err)
		pullCounterFailed.WithLabelValues(repository, reference).Inc()
		return
	}

	// the reference is tag, replace it with digest
	if _, err = digest.Parse(reference); err != nil {
		req = req.Clone(req.Context())
		req.URL.Path = strings.TrimSuffix(req.URL.Path, reference) + art.Digest
		req.URL.RawPath = req.URL.EscapedPath()
	}

	recorder := lib.NewResponseRecorder(w)
	proxy.ServeHTTP(recorder, req)
	// fire event, ignore the HEAD request and pulling request from replication service
	if !recorder.Success() || req.Method == http.MethodHead ||
		req.UserAgent() == registry.UserAgent {
		return
	}

	e := &metadata.PullArtifactEventMetadata{
		Artifact: &art.Artifact,
		Operator: operator.FromContext(req.Context()),
	}
	// the reference is tag
	if _, err = digest.Parse(reference); err != nil {
		e.Tag = reference
	}

	pullCounterSuccess.WithLabelValues(repository, reference).Inc()

	notification.AddEvent(req.Context(), e)
}

// just delete the artifact from database
func deleteManifest(w http.ResponseWriter, req *http.Request) {
	repository := router.Param(req.Context(), ":splat")
	reference := router.Param(req.Context(), ":reference")
	deleteCounterTotal.WithLabelValues(repository, reference).Inc()
	// v2 doesn't support delete by tag
	// add parse digest here is to return ErrDigestInvalidFormat before GetByReference throws an NOT_FOUND(404)
	// Do not add the logic into GetByReference as it's a shared method for PUT/GET/DELETE/Internal call,
	// and NOT_FOUND satisfy PUT/GET/Internal call.
	if _, err := digest.Parse(reference); err != nil {
		lib_http.SendError(w, errors.Wrapf(err, "unsupported digest %s", reference).WithCode(errors.UNSUPPORTED))
		deleteCounterFailed.WithLabelValues(repository, reference).Inc()
		return
	}
	art, err := artifact.Ctl.GetByReference(req.Context(), repository, reference, nil)
	if err != nil {
		lib_http.SendError(w, err)
		deleteCounterFailed.WithLabelValues(repository, reference).Inc()
		return
	}
	if err = artifact.Ctl.Delete(req.Context(), art.ID); err != nil {
		lib_http.SendError(w, err)
		deleteCounterFailed.WithLabelValues(repository, reference).Inc()
		return
	}
	deleteCounterSuccess.WithLabelValues(repository, reference).Inc()
	w.WriteHeader(http.StatusAccepted)
}

func putManifest(w http.ResponseWriter, req *http.Request) {
	repo := router.Param(req.Context(), ":splat")
	reference := router.Param(req.Context(), ":reference")
	pushCounterTotal.WithLabelValues(repo, reference).Inc()
	// make sure the repository exist before pushing the manifest
	_, _, err := repository.Ctl.Ensure(req.Context(), repo)
	if err != nil {
		lib_http.SendError(w, err)
		pushCounterFailed.WithLabelValues(repo, reference).Inc()
		return
	}

	buffer := lib.NewResponseBuffer(w)
	// proxy the req to the backend docker registry
	proxy.ServeHTTP(buffer, req)
	if !buffer.Success() {
		if _, err := buffer.Flush(); err != nil {
			log.Errorf("failed to flush: %v", err)
		}
		pushCounterFailed.WithLabelValues(repo, reference).Inc()
		return
	}

	// When got the response from the backend docker registry, the manifest and
	// tag are both ready, so we don't need to handle the issue anymore:
	// https://github.com/docker/distribution/issues/2625

	var tags []string
	dgt := reference
	// the reference is tag, get the digest from the response header
	if _, err = digest.Parse(reference); err != nil {
		dgt = buffer.Header().Get("Docker-Content-Digest")
		tags = append(tags, reference)
	}

	_, _, err = artifact.Ctl.Ensure(req.Context(), repo, dgt, tags...)
	if err != nil {
		lib_http.SendError(w, err)
		pushCounterFailed.WithLabelValues(repo, reference).Inc()
		return
	}

	// flush the origin response from the docker registry to the underlying response writer
	if _, err := buffer.Flush(); err != nil {
		log.Errorf("failed to flush: %v", err)
	}
	pushCounterSuccess.WithLabelValues(repo, reference).Inc()
}
