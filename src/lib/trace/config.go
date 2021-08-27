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
	"bytes"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/spf13/viper"
)

const (
	TraceEnvPrefix = "trace"
)

// C is the global configuration for trace
var C Config

func init() {
	viper.SetConfigType("json")
	viper.SetEnvPrefix(TraceEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	C = Config{Otel: OtelConfig{}, Jaeger: JaegerConfig{}}
	C.Enabled = viper.GetBool("enabled")
	C.SampleRate = viper.GetFloat64("sample_rate")
	C.Namespace = viper.GetString("namespace")
	C.ServiceName = viper.GetString("service_name")
	C.Jaeger.Endpoint = viper.GetString("jaeger_endpoint")
	C.Jaeger.Username = viper.GetString("jaeger_agent_username")
	C.Jaeger.Password = viper.GetString("jaeger_agent_password")
	C.Jaeger.AgentHost = viper.GetString("jaeger_agent_host")
	C.Jaeger.AgentPort = viper.GetString("jaeger_agent_port")
	C.Otel.Endpoint = viper.GetString("otel_endpoint")
	C.Otel.URLPath = viper.GetString("otel_url_path")
	C.Otel.Compression = viper.GetBool("otel_compression")
	C.Otel.Insecure = viper.GetBool("otel_insecure")
	C.Otel.Timeout = viper.GetInt("otel_timeout")
	var jsonExample = []byte(viper.GetString("attributes"))
	viper.ReadConfig(bytes.NewBuffer(jsonExample))
	fmt.Println(viper.GetStringMapString("attributes"))
	C.Attributes = viper.GetStringMapString("attributes")
	log.Infof("ns: %s attr %+v", C.Namespace, C.Attributes)
}

// OtelConfig is the configuration for otel
type OtelConfig struct {
	Endpoint    string `mapstructure:"otel_trace_endpoint"`
	URLPath     string `mapstructure:"otel_trace_url_path"`
	Compression bool   `mapstructure:"otel_trace_compression"`
	Insecure    bool   `mapstructure:"otel_trace_insecure"`
	Timeout     int    `mapstructure:"otel_trace_timeout"`
}

// JaegerConfig is the configuration for Jaeger
type JaegerConfig struct {
	Endpoint  string `mapstructure:"jaeger_endpoint"`
	Username  string `mapstructure:"jaeger_username"`
	Password  string `mapstructure:"jaeger_password"`
	AgentHost string `mapstructure:"jaeger_agent_host"`
	AgentPort string `mapstructure:"jaeger_agent_port"`
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

// GetConfig returns the global configuration for trace
func GetConfig() Config {
	return C
}

// Enabled returns whether trace is enabled
func Enabled() bool {
	return C.Enabled
}
