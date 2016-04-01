package compose

// Designed to support multiple applications
// currently support only one application
//
type ApplicationGraph struct {
	PrimaryApplications []Application // applications depends on others
}

const MAX_APP_SIZE = 100

func NewApplicationGraph(apps []Application) (*ApplicationGraph, error) {
	graph := &ApplicationGraph{
		PrimaryApplications: make([]Application, MAX_APP_SIZE),
	}
	for _, app := range apps {
		graph.PrimaryApplications = append(graph.PrimaryApplications, app)
	}

	return graph, nil
}
