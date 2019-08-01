package config

var (
	// Config is the configuration
	Config *Configuration
)

// Configuration holds the configuration information for notification
type Configuration struct {
	CoreURL          string
	CoreSecret       string
	TokenServiceURL  string
	JobserviceURL    string
	JobserviceSecret string
}
