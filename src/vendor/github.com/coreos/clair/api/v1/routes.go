// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/coreos/clair/api/context"
	"github.com/coreos/clair/database"
	"github.com/coreos/clair/utils"
	cerrors "github.com/coreos/clair/utils/errors"
	"github.com/coreos/clair/worker"
)

const (
	// These are the route identifiers for prometheus.
	postLayerRoute           = "v1/postLayer"
	getLayerRoute            = "v1/getLayer"
	deleteLayerRoute         = "v1/deleteLayer"
	getNamespacesRoute       = "v1/getNamespaces"
	getVulnerabilitiesRoute  = "v1/getVulnerabilities"
	postVulnerabilityRoute   = "v1/postVulnerability"
	getVulnerabilityRoute    = "v1/getVulnerability"
	putVulnerabilityRoute    = "v1/putVulnerability"
	deleteVulnerabilityRoute = "v1/deleteVulnerability"
	getFixesRoute            = "v1/getFixes"
	putFixRoute              = "v1/putFix"
	deleteFixRoute           = "v1/deleteFix"
	getNotificationRoute     = "v1/getNotification"
	deleteNotificationRoute  = "v1/deleteNotification"
	getMetricsRoute          = "v1/getMetrics"

	// maxBodySize restricts client request bodies to 1MiB.
	maxBodySize int64 = 1048576

	// statusUnprocessableEntity represents the 422 (Unprocessable Entity) status code, which means
	// the server understands the content type of the request entity
	// (hence a 415(Unsupported Media Type) status code is inappropriate), and the syntax of the
	// request entity is correct (thus a 400 (Bad Request) status code is inappropriate) but was
	// unable to process the contained instructions.
	statusUnprocessableEntity = 422
)

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, maxBodySize)).Decode(v)
}

func writeResponse(w http.ResponseWriter, r *http.Request, status int, resp interface{}) {
	// Headers must be written before the response.
	header := w.Header()
	header.Set("Content-Type", "application/json;charset=utf-8")
	header.Set("Server", "clair")

	// Gzip the response if the client supports it.
	var writer io.Writer = w
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()
		writer = gzipWriter

		header.Set("Content-Encoding", "gzip")
	}

	// Write the response.
	w.WriteHeader(status)
	err := json.NewEncoder(writer).Encode(resp)

	if err != nil {
		switch err.(type) {
		case *json.MarshalerError, *json.UnsupportedTypeError, *json.UnsupportedValueError:
			panic("v1: failed to marshal response: " + err.Error())
		default:
			log.Warningf("failed to write response: %s", err.Error())
		}
	}
}

func postLayer(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	request := LayerEnvelope{}
	err := decodeJSON(r, &request)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, LayerEnvelope{Error: &Error{err.Error()}})
		return postLayerRoute, http.StatusBadRequest
	}

	if request.Layer == nil {
		writeResponse(w, r, http.StatusBadRequest, LayerEnvelope{Error: &Error{"failed to provide layer"}})
		return postLayerRoute, http.StatusBadRequest
	}

	err = worker.Process(ctx.Store, request.Layer.Format, request.Layer.Name, request.Layer.ParentName, request.Layer.Path, request.Layer.Headers)
	if err != nil {
		if err == utils.ErrCouldNotExtract ||
			err == utils.ErrExtractedFileTooBig ||
			err == worker.ErrUnsupported {
			writeResponse(w, r, statusUnprocessableEntity, LayerEnvelope{Error: &Error{err.Error()}})
			return postLayerRoute, statusUnprocessableEntity
		}

		if _, badreq := err.(*cerrors.ErrBadRequest); badreq {
			writeResponse(w, r, http.StatusBadRequest, LayerEnvelope{Error: &Error{err.Error()}})
			return postLayerRoute, http.StatusBadRequest
		}

		writeResponse(w, r, http.StatusInternalServerError, LayerEnvelope{Error: &Error{err.Error()}})
		return postLayerRoute, http.StatusInternalServerError
	}

	writeResponse(w, r, http.StatusCreated, LayerEnvelope{Layer: &Layer{
		Name:             request.Layer.Name,
		ParentName:       request.Layer.ParentName,
		Path:             request.Layer.Path,
		Headers:          request.Layer.Headers,
		Format:           request.Layer.Format,
		IndexedByVersion: worker.Version,
	}})
	return postLayerRoute, http.StatusCreated
}

func getLayer(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	_, withFeatures := r.URL.Query()["features"]
	_, withVulnerabilities := r.URL.Query()["vulnerabilities"]

	dbLayer, err := ctx.Store.FindLayer(p.ByName("layerName"), withFeatures, withVulnerabilities)
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, LayerEnvelope{Error: &Error{err.Error()}})
		return getLayerRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, LayerEnvelope{Error: &Error{err.Error()}})
		return getLayerRoute, http.StatusInternalServerError
	}

	layer := LayerFromDatabaseModel(dbLayer, withFeatures, withVulnerabilities)

	writeResponse(w, r, http.StatusOK, LayerEnvelope{Layer: &layer})
	return getLayerRoute, http.StatusOK
}

func deleteLayer(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	err := ctx.Store.DeleteLayer(p.ByName("layerName"))
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, LayerEnvelope{Error: &Error{err.Error()}})
		return deleteLayerRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, LayerEnvelope{Error: &Error{err.Error()}})
		return deleteLayerRoute, http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusOK)
	return deleteLayerRoute, http.StatusOK
}

func getNamespaces(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	dbNamespaces, err := ctx.Store.ListNamespaces()
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, NamespaceEnvelope{Error: &Error{err.Error()}})
		return getNamespacesRoute, http.StatusInternalServerError
	}
	var namespaces []Namespace
	for _, dbNamespace := range dbNamespaces {
		namespaces = append(namespaces, Namespace{Name: dbNamespace.Name})
	}

	writeResponse(w, r, http.StatusOK, NamespaceEnvelope{Namespaces: &namespaces})
	return getNamespacesRoute, http.StatusOK
}

func getVulnerabilities(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	query := r.URL.Query()

	limitStrs, limitExists := query["limit"]
	if !limitExists {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"must provide limit query parameter"}})
		return getVulnerabilitiesRoute, http.StatusBadRequest
	}
	limit, err := strconv.Atoi(limitStrs[0])
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"invalid limit format: " + err.Error()}})
		return getVulnerabilitiesRoute, http.StatusBadRequest
	} else if limit < 0 {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"limit value should not be less than zero"}})
		return getVulnerabilitiesRoute, http.StatusBadRequest
	}

	page := 0
	pageStrs, pageExists := query["page"]
	if pageExists {
		err = tokenUnmarshal(pageStrs[0], ctx.Config.PaginationKey, &page)
		if err != nil {
			writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"invalid page format: " + err.Error()}})
			return getNotificationRoute, http.StatusBadRequest
		}
	}

	namespace := p.ByName("namespaceName")
	if namespace == "" {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"namespace should not be empty"}})
		return getNotificationRoute, http.StatusBadRequest
	}

	dbVulns, nextPage, err := ctx.Store.ListVulnerabilities(namespace, limit, page)
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return getVulnerabilityRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return getVulnerabilitiesRoute, http.StatusInternalServerError
	}

	var vulns []Vulnerability
	for _, dbVuln := range dbVulns {
		vuln := VulnerabilityFromDatabaseModel(dbVuln, false)
		vulns = append(vulns, vuln)
	}

	var nextPageStr string
	if nextPage != -1 {
		nextPageBytes, err := tokenMarshal(nextPage, ctx.Config.PaginationKey)
		if err != nil {
			writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"failed to marshal token: " + err.Error()}})
			return getNotificationRoute, http.StatusBadRequest
		}
		nextPageStr = string(nextPageBytes)
	}

	writeResponse(w, r, http.StatusOK, VulnerabilityEnvelope{Vulnerabilities: &vulns, NextPage: nextPageStr})
	return getVulnerabilitiesRoute, http.StatusOK
}

func postVulnerability(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	request := VulnerabilityEnvelope{}
	err := decodeJSON(r, &request)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return postVulnerabilityRoute, http.StatusBadRequest
	}

	if request.Vulnerability == nil {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"failed to provide vulnerability"}})
		return postVulnerabilityRoute, http.StatusBadRequest
	}

	vuln, err := request.Vulnerability.DatabaseModel()
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return postVulnerabilityRoute, http.StatusBadRequest
	}

	err = ctx.Store.InsertVulnerabilities([]database.Vulnerability{vuln}, true)
	if err != nil {
		switch err.(type) {
		case *cerrors.ErrBadRequest:
			writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{err.Error()}})
			return postVulnerabilityRoute, http.StatusBadRequest
		default:
			writeResponse(w, r, http.StatusInternalServerError, VulnerabilityEnvelope{Error: &Error{err.Error()}})
			return postVulnerabilityRoute, http.StatusInternalServerError
		}
	}

	writeResponse(w, r, http.StatusCreated, VulnerabilityEnvelope{Vulnerability: request.Vulnerability})
	return postVulnerabilityRoute, http.StatusCreated
}

func getVulnerability(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	_, withFixedIn := r.URL.Query()["fixedIn"]

	dbVuln, err := ctx.Store.FindVulnerability(p.ByName("namespaceName"), p.ByName("vulnerabilityName"))
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return getVulnerabilityRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return getVulnerabilityRoute, http.StatusInternalServerError
	}

	vuln := VulnerabilityFromDatabaseModel(dbVuln, withFixedIn)

	writeResponse(w, r, http.StatusOK, VulnerabilityEnvelope{Vulnerability: &vuln})
	return getVulnerabilityRoute, http.StatusOK
}

func putVulnerability(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	request := VulnerabilityEnvelope{}
	err := decodeJSON(r, &request)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return putVulnerabilityRoute, http.StatusBadRequest
	}

	if request.Vulnerability == nil {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"failed to provide vulnerability"}})
		return putVulnerabilityRoute, http.StatusBadRequest
	}

	if len(request.Vulnerability.FixedIn) != 0 {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{"Vulnerability.FixedIn must be empty"}})
		return putVulnerabilityRoute, http.StatusBadRequest
	}

	vuln, err := request.Vulnerability.DatabaseModel()
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return putVulnerabilityRoute, http.StatusBadRequest
	}

	vuln.Namespace.Name = p.ByName("namespaceName")
	vuln.Name = p.ByName("vulnerabilityName")

	err = ctx.Store.InsertVulnerabilities([]database.Vulnerability{vuln}, true)
	if err != nil {
		switch err.(type) {
		case *cerrors.ErrBadRequest:
			writeResponse(w, r, http.StatusBadRequest, VulnerabilityEnvelope{Error: &Error{err.Error()}})
			return putVulnerabilityRoute, http.StatusBadRequest
		default:
			writeResponse(w, r, http.StatusInternalServerError, VulnerabilityEnvelope{Error: &Error{err.Error()}})
			return putVulnerabilityRoute, http.StatusInternalServerError
		}
	}

	writeResponse(w, r, http.StatusOK, VulnerabilityEnvelope{Vulnerability: request.Vulnerability})
	return putVulnerabilityRoute, http.StatusOK
}

func deleteVulnerability(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	err := ctx.Store.DeleteVulnerability(p.ByName("namespaceName"), p.ByName("vulnerabilityName"))
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return deleteVulnerabilityRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, VulnerabilityEnvelope{Error: &Error{err.Error()}})
		return deleteVulnerabilityRoute, http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusOK)
	return deleteVulnerabilityRoute, http.StatusOK
}

func getFixes(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	dbVuln, err := ctx.Store.FindVulnerability(p.ByName("namespaceName"), p.ByName("vulnerabilityName"))
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, FeatureEnvelope{Error: &Error{err.Error()}})
		return getFixesRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, FeatureEnvelope{Error: &Error{err.Error()}})
		return getFixesRoute, http.StatusInternalServerError
	}

	vuln := VulnerabilityFromDatabaseModel(dbVuln, true)
	writeResponse(w, r, http.StatusOK, FeatureEnvelope{Features: &vuln.FixedIn})
	return getFixesRoute, http.StatusOK
}

func putFix(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	request := FeatureEnvelope{}
	err := decodeJSON(r, &request)
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, FeatureEnvelope{Error: &Error{err.Error()}})
		return putFixRoute, http.StatusBadRequest
	}

	if request.Feature == nil {
		writeResponse(w, r, http.StatusBadRequest, FeatureEnvelope{Error: &Error{"failed to provide feature"}})
		return putFixRoute, http.StatusBadRequest
	}

	if request.Feature.Name != p.ByName("fixName") {
		writeResponse(w, r, http.StatusBadRequest, FeatureEnvelope{Error: &Error{"feature name in URL and JSON do not match"}})
		return putFixRoute, http.StatusBadRequest
	}

	dbFix, err := request.Feature.DatabaseModel()
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, FeatureEnvelope{Error: &Error{err.Error()}})
		return putFixRoute, http.StatusBadRequest
	}

	err = ctx.Store.InsertVulnerabilityFixes(p.ByName("vulnerabilityNamespace"), p.ByName("vulnerabilityName"), []database.FeatureVersion{dbFix})
	if err != nil {
		switch err.(type) {
		case *cerrors.ErrBadRequest:
			writeResponse(w, r, http.StatusBadRequest, FeatureEnvelope{Error: &Error{err.Error()}})
			return putFixRoute, http.StatusBadRequest
		default:
			if err == cerrors.ErrNotFound {
				writeResponse(w, r, http.StatusNotFound, FeatureEnvelope{Error: &Error{err.Error()}})
				return putFixRoute, http.StatusNotFound
			}
			writeResponse(w, r, http.StatusInternalServerError, FeatureEnvelope{Error: &Error{err.Error()}})
			return putFixRoute, http.StatusInternalServerError
		}
	}

	writeResponse(w, r, http.StatusOK, FeatureEnvelope{Feature: request.Feature})
	return putFixRoute, http.StatusOK
}

func deleteFix(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	err := ctx.Store.DeleteVulnerabilityFix(p.ByName("vulnerabilityNamespace"), p.ByName("vulnerabilityName"), p.ByName("fixName"))
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, FeatureEnvelope{Error: &Error{err.Error()}})
		return deleteFixRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, FeatureEnvelope{Error: &Error{err.Error()}})
		return deleteFixRoute, http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusOK)
	return deleteFixRoute, http.StatusOK
}

func getNotification(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	query := r.URL.Query()

	limitStrs, limitExists := query["limit"]
	if !limitExists {
		writeResponse(w, r, http.StatusBadRequest, NotificationEnvelope{Error: &Error{"must provide limit query parameter"}})
		return getNotificationRoute, http.StatusBadRequest
	}
	limit, err := strconv.Atoi(limitStrs[0])
	if err != nil {
		writeResponse(w, r, http.StatusBadRequest, NotificationEnvelope{Error: &Error{"invalid limit format: " + err.Error()}})
		return getNotificationRoute, http.StatusBadRequest
	}

	var pageToken string
	page := database.VulnerabilityNotificationFirstPage
	pageStrs, pageExists := query["page"]
	if pageExists {
		err := tokenUnmarshal(pageStrs[0], ctx.Config.PaginationKey, &page)
		if err != nil {
			writeResponse(w, r, http.StatusBadRequest, NotificationEnvelope{Error: &Error{"invalid page format: " + err.Error()}})
			return getNotificationRoute, http.StatusBadRequest
		}
		pageToken = pageStrs[0]
	} else {
		pageTokenBytes, err := tokenMarshal(page, ctx.Config.PaginationKey)
		if err != nil {
			writeResponse(w, r, http.StatusBadRequest, NotificationEnvelope{Error: &Error{"failed to marshal token: " + err.Error()}})
			return getNotificationRoute, http.StatusBadRequest
		}
		pageToken = string(pageTokenBytes)
	}

	dbNotification, nextPage, err := ctx.Store.GetNotification(p.ByName("notificationName"), limit, page)
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, NotificationEnvelope{Error: &Error{err.Error()}})
		return deleteNotificationRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, NotificationEnvelope{Error: &Error{err.Error()}})
		return getNotificationRoute, http.StatusInternalServerError
	}

	notification := NotificationFromDatabaseModel(dbNotification, limit, pageToken, nextPage, ctx.Config.PaginationKey)

	writeResponse(w, r, http.StatusOK, NotificationEnvelope{Notification: &notification})
	return getNotificationRoute, http.StatusOK
}

func deleteNotification(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	err := ctx.Store.DeleteNotification(p.ByName("notificationName"))
	if err == cerrors.ErrNotFound {
		writeResponse(w, r, http.StatusNotFound, NotificationEnvelope{Error: &Error{err.Error()}})
		return deleteNotificationRoute, http.StatusNotFound
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, NotificationEnvelope{Error: &Error{err.Error()}})
		return deleteNotificationRoute, http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusOK)
	return deleteNotificationRoute, http.StatusOK
}

func getMetrics(w http.ResponseWriter, r *http.Request, p httprouter.Params, ctx *context.RouteContext) (string, int) {
	prometheus.Handler().ServeHTTP(w, r)
	return getMetricsRoute, 0
}
