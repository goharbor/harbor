package compose

type Application struct {
	IsPrimary   bool    // application depends on other applications
	MeetCritia  bool    // application running now meet critia specified by compose file
	Name        string  `json: "name" yaml: "name"`
	Image       string  `json: "image" yaml: "image"`
	Cmd         string  `json: "cmd" yaml: "cmd"`
	Cpu         float32 `json: "cpu" yaml: "cpu"`
	Mem         float32 `json: "mem" yaml: "mem"`
	Environment []Environment
	Labels      []Label
	Volumn      []Volumn
	Port        []Port

	Dependencies []*Application
}

type Environment struct {
	Key   string
	Value string
}

type Label struct {
	Key   string
	Value string
}

type Volumn struct {
	Container  string
	Host       string
	Permission string
}

type Port struct {
	HostAddr      string
	HostPort      int
	ContaienrAddr string
	ContaienrPort int
	Protocol      string
}
