package gock

// Options represents customized option for gock
type Options struct {
	// DisableRegexpHost stores if the host is only a plain string rather than regular expression,
	// if DisableRegexpHost is true, host sets in gock.New(...) will be treated as plain string
	DisableRegexpHost bool
}
