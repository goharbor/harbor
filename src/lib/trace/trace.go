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
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	service          = "core"
	environment      = "production"
	traceServiceName = "goharbor/harbor"
)

type ProviderConfig struct {
	ExporterType string
	URL          string
	Attribute    map[string]string
}

func initExporter(ctx context.Context) (tracesdk.SpanExporter, error) {
	var err error
	var exp tracesdk.SpanExporter
	cfg := GetConfig()
	if len(cfg.Jaeger.Endpoint) != 0 {
		// Jaeger collector exporter
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(cfg.Jaeger.Endpoint),
			jaeger.WithUsername(cfg.Jaeger.Username),
			jaeger.WithPassword(cfg.Jaeger.Password),
		))
	} else if len(cfg.Jaeger.AgentHost) != 0 {
		// Jaeger agent exporter
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.Jaeger.AgentHost),
			jaeger.WithAgentPort(cfg.Jaeger.AgentPort),
		))
	} else if len(cfg.Otel.Endpoint) != 0 {
		// Otel exporter
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Otel.Endpoint),
			otlptracehttp.WithURLPath(cfg.Otel.URLPath),
			otlptracehttp.WithTimeout(time.Duration(cfg.Otel.Timeout) * time.Second),
		}
		if cfg.Otel.Compression {
			opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
		}

		exp, err = otlptracehttp.New(ctx, opts...)
	} else {
		log.Fatalf("Trace is enabled but no tracer provider is specified")
	}
	return exp, err
}

func initProvider(exp tracesdk.SpanExporter) (*tracesdk.TracerProvider, error) {
	cfg := GetConfig()
	ops := make([]tracesdk.TracerProviderOption, 4)

	ops = append(ops,
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
		)),
	)

	attriSlice := make([]attribute.KeyValue, 0, len(cfg.Attribute))
	if cfg.Attribute != nil {
		for i, a := range cfg.Attribute {
			attriSlice = append(attriSlice, attribute.String(i, a))
		}
		ops = append(ops, tracesdk.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attriSlice...)))
	}
	bsp := tracesdk.NewBatchSpanProcessor(exp)
	ops = append(ops, tracesdk.WithSpanProcessor(bsp), tracesdk.WithSampler(tracesdk.TraceIDRatioBased(cfg.SampleRate)))
	tp := tracesdk.NewTracerProvider(ops...)

	return tp, nil
}

func InitGlobalTracer(ctx context.Context) *tracesdk.TracerProvider {
	exp, err := initExporter(ctx)
	handleErr(err, "fail in exporter initialization")
	tp, err := initProvider(exp)
	handleErr(err, "fail in tracer initialization")
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tp)
	return tp
}

func GetGlobalTracer(instrumentationName string, opts ...trace.TracerOption) trace.Tracer {
	return otel.GetTracerProvider().Tracer(instrumentationName, opts...)
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
