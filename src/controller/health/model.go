package health

// OverallHealthStatus defines the overall health status of the system
type OverallHealthStatus struct {
	Status     string                   `json:"status"`
	Components []*ComponentHealthStatus `json:"components"`
}

// ComponentHealthStatus defines the specific component health status
type ComponentHealthStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthy bool

func (h healthy) String() string {
	if h {
		return "healthy"
	}
	return "unhealthy"
}
