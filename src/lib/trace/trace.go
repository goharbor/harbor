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

package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/version"
)

func initExporter(ctx context.Context) (tracesdk.SpanExporter, error) {
	var err error
	var exp tracesdk.SpanExporter
	cfg := GetGlobalConfig()
	if len(cfg.Jaeger.Endpoint) != 0 {
		// Jaeger collector exporter
		log.Infof("init trace provider jaeger collector on %s with user %s", cfg.Jaeger.Endpoint, cfg.Jaeger.Username)
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(cfg.Jaeger.Endpoint),
			jaeger.WithUsername(cfg.Jaeger.Username),
			jaeger.WithPassword(cfg.Jaeger.Password),
		))
	} else if len(cfg.Jaeger.AgentHost) != 0 {
		// Jaeger agent exporter
		log.Infof("init trace provider jaeger agent on %s", cfg.Jaeger.AgentHost)
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.Jaeger.AgentHost),
			jaeger.WithAgentPort(cfg.Jaeger.AgentPort),
		))
	} else if len(cfg.Otel.Endpoint) != 0 {
		// Otel exporter
		log.Infof("init trace provider otel on %s/%s", cfg.Otel.Endpoint, cfg.Otel.URLPath)
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Otel.Endpoint),
			otlptracehttp.WithURLPath(cfg.Otel.URLPath),
			otlptracehttp.WithTimeout(time.Duration(cfg.Otel.Timeout) * time.Second),
		}
		if cfg.Otel.Compression {
			opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
		}
		if cfg.Otel.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exp, err = otlptracehttp.New(ctx, opts...)
	} else {
		log.Fatalf("Trace is enabled but no tracer provider is specified")
	}
	return exp, err
}

func initProvider(exp tracesdk.SpanExporter) *tracesdk.TracerProvider {
	cfg := GetGlobalConfig()

	// prepare attribute resources
	attriSlice := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
	}
	if len(version.ReleaseVersion) != 0 {
		attriSlice = append(attriSlice, semconv.ServiceVersionKey.String(version.ReleaseVersion))
	}
	if cfg.Namespace != "" {
		attriSlice = append(attriSlice, semconv.ServiceNamespaceKey.String(cfg.Namespace))
	}
	if cfg.Attributes != nil {
		for i, a := range cfg.Attributes {
			attriSlice = append(attriSlice, attribute.String(i, a))
		}
	}

	// prepare tp options
	ops := make([]tracesdk.TracerProviderOption, 0, 4)
	ops = append(ops,
		// Always be sure to batch in production.
		// tracesdk.WithBatcher(exp),
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attriSlice...)),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(cfg.SampleRate)),
	)
	// init trace provider
	return tracesdk.NewTracerProvider(ops...)
}

// ShutdownFunc is a function to shutdown the trace provider
type ShutdownFunc func()

// Shutdown shutdown the trace provider
func (s ShutdownFunc) Shutdown() {
	s()
}

// Init initializes the trace provider
func InitGlobalTracer(ctx context.Context) ShutdownFunc {
	if !Enabled() {
		otel.SetTracerProvider(oteltrace.NewNoopTracerProvider())
		return func() {}
	}
	exp, err := initExporter(ctx)
	handleErr(err, "fail in exporter initialization")
	tp := initProvider(exp)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return func() {
		log.Infof("shutdown trace provider")
		handleErr(tp.Shutdown(ctx), "fail in tracer shutdown")
	}
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
