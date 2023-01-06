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
	"fmt"
)

const (
	// TraceEnvPrefix is the prefix of trace related env.
	TraceEnvPrefix = "trace"
)

// C is the global configuration for trace
var C Config

// InitGlobalConfig inits global config.
func InitGlobalConfig(opts ...Option) {
	C = NewConfig(opts...)
}

// OtelConfig is the configuration for otel
type OtelConfig struct {
	Endpoint    string `mapstructure:"otel_trace_endpoint"`
	URLPath     string `mapstructure:"otel_trace_url_path"`
	Compression bool   `mapstructure:"otel_trace_compression"`
	Insecure    bool   `mapstructure:"otel_trace_insecure"`
	Timeout     int    `mapstructure:"otel_trace_timeout"`
}

func (c *OtelConfig) String() string {
	return fmt.Sprintf("endpoint: %s, url_path: %s, compression: %t, insecure: %t, timeout: %d",
		c.Endpoint, c.URLPath, c.Compression, c.Insecure, c.Timeout)
}

// JaegerConfig is the configuration for Jaeger
type JaegerConfig struct {
	Endpoint  string `mapstructure:"jaeger_endpoint"`
	Username  string `mapstructure:"jaeger_username"`
	Password  string `mapstructure:"jaeger_password"`
	AgentHost string `mapstructure:"jaeger_agent_host"`
	AgentPort string `mapstructure:"jaeger_agent_port"`
}

func (c *JaegerConfig) String() string {
	return fmt.Sprintf("endpoint: %s, username: %s, password: %s, agent_host: %s, agent_port: %s",
		c.Endpoint, c.Username, c.Password, c.AgentHost, c.AgentPort)
}

// Config is the configuration for trace
type Config struct {
	Enabled     bool    `mapstructure:"enabled"`
	SampleRate  float64 `mapstructure:"sample_rate"`
	Namespace   string  `mapstructure:"namespace"`
	ServiceName string  `mapstructure:"service_name"`
	Jaeger      JaegerConfig
	Otel        OtelConfig
	Attributes  map[string]string
}

func (c *Config) String() string {
	return fmt.Sprintf("{Enabled: %v, ServiceName: %v,  SampleRate: %v, Namespace: %v, ServiceName: %v, Jaeger: %v, Otel: %v}", c.Enabled, c.ServiceName, c.SampleRate, c.Namespace, c.ServiceName, c.Jaeger, c.Otel)
}

// Option is the wrapper for changing config.
type Option func(*Config)

// WithEnabled pass in the enabled.
func WithEnabled(enabled bool) Option {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithSampleRate pass in the sample rate.
func WithSampleRate(sampleRate float64) Option {
	return func(c *Config) {
		c.SampleRate = sampleRate
	}
}

// WithNamespace pass in the namespace.
func WithNamespace(namespace string) Option {
	return func(c *Config) {
		c.Namespace = namespace
	}
}

// WithServiceName pass in the service name.
func WithServiceName(serviceName string) Option {
	return func(c *Config) {
		c.ServiceName = serviceName
	}
}

// WithAttributes pass in the attributes.
func WithAttributes(attributes map[string]string) Option {
	return func(c *Config) {
		c.Attributes = attributes
	}
}

// WithJaegerEndpoint pass in the jaeger endpoint.
func WithJaegerEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.Jaeger.Endpoint = endpoint
	}
}

// WithJaegerUsername pass in the jaeger username.
func WithJaegerUsername(username string) Option {
	return func(c *Config) {
		c.Jaeger.Username = username
	}
}

// WithJaegerPassword pass in the jaeger password.
func WithJaegerPassword(password string) Option {
	return func(c *Config) {
		c.Jaeger.Password = password
	}
}

// WithJaegerAgentHost pass in the jaeger agent host.
func WithJaegerAgentHost(host string) Option {
	return func(c *Config) {
		c.Jaeger.AgentHost = host
	}
}

// WithJaegerAgentPort pass in the jaeger agent port.
func WithJaegerAgentPort(port string) Option {
	return func(c *Config) {
		c.Jaeger.AgentPort = port
	}
}

// WithOtelEndpoint pass in the otel endpoint.
func WithOtelEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.Otel.Endpoint = endpoint
	}
}

// WithOtelURLPath pass in the url path.
func WithOtelURLPath(urlPath string) Option {
	return func(c *Config) {
		c.Otel.URLPath = urlPath
	}
}

// WithOtelCompression pass in the otel compression.
func WithOtelCompression(compression bool) Option {
	return func(c *Config) {
		c.Otel.Compression = compression
	}
}

// WithOtelInsecure pass in the otel insecure.
func WithOtelInsecure(insecure bool) Option {
	return func(c *Config) {
		c.Otel.Insecure = insecure
	}
}

// WithOtelTimeout pass in the otel timeout.
func WithOtelTimeout(timeout int) Option {
	return func(c *Config) {
		c.Otel.Timeout = timeout
	}
}

// NewConfig returns config which generated by passed in options.
func NewConfig(opts ...Option) Config {
	c := Config{Otel: OtelConfig{}, Jaeger: JaegerConfig{}}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// GetGlobalConfig returns the global configuration for trace
func GetGlobalConfig() Config {
	return C
}

// Enabled returns whether trace is enabled
func Enabled() bool {
	return C.Enabled
}
