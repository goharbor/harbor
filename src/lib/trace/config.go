package trace

import (
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/spf13/viper"
)

const (
	TraceEnvPrefix = "Trace"
)

var C Config

func init() {
	log.Infof("parsing trace config from env")
	viper.AutomaticEnv()
	viper.SetEnvPrefix(TraceEnvPrefix)
	log.Infof("Set env prefix to %s", TraceEnvPrefix)
	var config Config
	var jaeger JaegerConfig
	var otel OtelConfig
	viper.Unmarshal(&config)
	viper.Unmarshal(&jaeger)
	viper.Unmarshal(&otel)
	config.Jaeger = jaeger
	config.Otel = otel
	C = config
}

type OtelTraceConfig struct {
}

type OtelConfig struct {
	Endpoint    string `mapstructure:"otel_trace_endpoint"`
	URLPath     string `mapstructure:"otel_trace_url_path"`
	Compression bool   `mapstructure:"otel_trace_compression"`
	Insecure    bool   `mapstructure:"otel_trace_insecure"`
	Timeout     int    `mapstructure:"otel_trace_timeout"`
}
type JaegerConfig struct {
	Endpoint  string `mapstructure:"jaeger_endpoint"`
	Username  string `mapstructure:"jaeger_username"`
	Password  string `mapstructure:"jaeger_password"`
	AgentHost string `mapstructure:"jaeger_agent_host"`
	AgentPort string `mapstructure:"jaeger_agent_port"`
}
type Config struct {
	Enabled    bool    `mapstructure:"enabled"`
	SampleRate float64 `mapstructure:"sample_rate"`
	Jaeger     JaegerConfig
	Otel       OtelConfig
	Attribute  map[string]string
}

func GetConfig() Config {
	return C
}

func Enabled() bool {
	return C.Enabled
}
