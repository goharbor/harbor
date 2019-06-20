package middlewares

// const variables
const (
	READONLY         = "readonly"
	URL              = "url"
	MUITIPLEMANIFEST = "manifest"
	LISTREPO         = "listrepo"
	CONTENTTRUST     = "contenttrust"
	VULNERABLE       = "vulnerable"
	REGQUOTA         = "regquota"
	BLOBQUOTA        = "blobquota"
)

// sequential organization
var Middlewares = []string{READONLY, URL, MUITIPLEMANIFEST, LISTREPO, CONTENTTRUST, VULNERABLE, BLOBQUOTA, REGQUOTA}
