package envs

// ConcourseCILdapEnv : Ldap env for concourse pipeline
var ConcourseCILdapEnv = Environment{
	Protocol:       "https",
	TestingProject: "concoursecitesting01",
	ImageName:      "busybox",
	ImageTag:       "latest",
	CAFile:         "../../../ca.crt",
	KeyFile:        "../../../key.crt",
	CertFile:       "../../../cert.crt",
	Account:        "mike",
	Password:       "zhu88jie",
	Admin:          "admin",
	AdminPass:      "pksxgxmifc0cnwa5px9h",
	Hostname:       "30.0.0.3",
}
