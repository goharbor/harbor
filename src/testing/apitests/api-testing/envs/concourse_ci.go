package envs

// ConcourseCIEnv : Env for concourse pipeline
var ConcourseCIEnv = Environment{
	Protocol:       "https",
	TestingProject: "concoursecitesting01",
	ImageName:      "busybox",
	ImageTag:       "latest",
	CAFile:         "../../../ca.crt",
	KeyFile:        "../../../key.crt",
	CertFile:       "../../../cert.crt",
	Account:        "cody",
	Password:       "Admin!23",
	Admin:          "admin",
	AdminPass:      "pksxgxmifc0cnwa5px9h",
	Hostname:       "10.112.122.1",
}
