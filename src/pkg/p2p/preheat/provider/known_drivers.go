package provider

// knownDrivers is static driver Factory registry
var knownDrivers = map[string]Factory{
	"dragonfly": DragonflyFactory,
	"kraken":    KrakenFactory,
}

// ListProviders returns all the registered drivers.
func ListProviders() ([]*Metadata, error) {
	results := []*Metadata{}

	for _, f := range knownDrivers {
		drv, err := f(nil)
		if err != nil {
			return nil, err
		}

		results = append(results, drv.Self())
	}

	return results, nil
}

// GetProvider returns the driver factory identified by the ID.
//
// If exists, bool flag will be set to be true and a non-nil reference will be returned.
func GetProvider(ID string) (Factory, bool) {
	f, ok := knownDrivers[ID]

	return f, ok
}
