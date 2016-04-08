package channel

type ChannelHttpConfig struct {
	Password  string
	Principle string
	Token     string
	Cert      string
	Pem       string
	Type      string //token, http_basic, pk
	AppApiUrl string
}
