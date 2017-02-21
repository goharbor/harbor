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

package context

import (
	"net/http"
	"strconv"
	"time"

	"github.com/coreos/pkg/capnslog"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/coreos/clair/config"
	"github.com/coreos/clair/database"
	"github.com/coreos/clair/utils"
)

var (
	log = capnslog.NewPackageLogger("github.com/coreos/clair", "api")

	promResponseDurationMilliseconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "clair_api_response_duration_milliseconds",
		Help:    "The duration of time it takes to receieve and write a response to an API request",
		Buckets: prometheus.ExponentialBuckets(9.375, 2, 10),
	}, []string{"route", "code"})
)

func init() {
	prometheus.MustRegister(promResponseDurationMilliseconds)
}

type Handler func(http.ResponseWriter, *http.Request, httprouter.Params, *RouteContext) (route string, status int)

func HTTPHandler(handler Handler, ctx *RouteContext) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()
		route, status := handler(w, r, p, ctx)
		statusStr := strconv.Itoa(status)
		if status == 0 {
			statusStr = "???"
		}
		utils.PrometheusObserveTimeMilliseconds(promResponseDurationMilliseconds.WithLabelValues(route, statusStr), start)

		log.Infof("%s \"%s %s\" %s (%s)", r.RemoteAddr, r.Method, r.RequestURI, statusStr, time.Since(start))
	}
}

type RouteContext struct {
	Store  database.Datastore
	Config *config.APIConfig
}
