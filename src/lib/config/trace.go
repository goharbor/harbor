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

package config

import (
	"context"

	"github.com/goharbor/harbor/src/common"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
)

func InitTraceConfig(ctx context.Context) {
	cfgMgr, err := GetManager(common.InMemoryCfgManager)
	if err != nil {
		log.Fatalf("failed to get config manager: %v", err)
	}
	if cfgMgr.Get(ctx, common.TraceEnabled).GetBool() {
		tracelib.InitGlobalConfig(
			tracelib.WithEnabled(true),
			tracelib.WithServiceName(cfgMgr.Get(ctx, common.TraceServiceName).GetString()),
			tracelib.WithNamespace(cfgMgr.Get(ctx, common.TraceNamespace).GetString()),
			tracelib.WithSampleRate(cfgMgr.Get(ctx, common.TraceSampleRate).GetFloat64()),
			tracelib.WithAttributes(cfgMgr.Get(ctx, common.TraceAttributes).GetStringToStringMap()),
			tracelib.WithJaegerEndpoint(cfgMgr.Get(ctx, common.TraceJaegerEndpoint).GetString()),
			tracelib.WithJaegerUsername(cfgMgr.Get(ctx, common.TraceJaegerUsername).GetString()),
			tracelib.WithJaegerPassword(cfgMgr.Get(ctx, common.TraceJaegerPassword).GetString()),
			tracelib.WithJaegerAgentHost(cfgMgr.Get(ctx, common.TraceJaegerAgentHost).GetString()),
			tracelib.WithJaegerAgentPort(cfgMgr.Get(ctx, common.TraceJaegerAgentPort).GetString()),
			tracelib.WithOtelEndpoint(cfgMgr.Get(ctx, common.TraceOtelEndpoint).GetString()),
			tracelib.WithOtelURLPath(cfgMgr.Get(ctx, common.TraceOtelURLPath).GetString()),
			tracelib.WithOtelCompression(cfgMgr.Get(ctx, common.TraceOtelCompression).GetBool()),
			tracelib.WithOtelInsecure(cfgMgr.Get(ctx, common.TraceOtelInsecure).GetBool()),
			tracelib.WithOtelTimeout(cfgMgr.Get(ctx, common.TraceOtelTimeout).GetInt()),
		)
		commonhttp.AddTracingWithGlobalTransport()
	}
}
