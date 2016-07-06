package servie

import (
	"os"
)

type Config struct {
	Workspace   string
	Project     string
	Imagename   string
	Gituri      string
	Storemethod string
	Branch      string
}

var config Config

func Pairs() Config {
	return config
}
func InitParam() {
	initConfig(os.Getenv("HARBOR_GIT_WORKSPACE"), os.Getenv("HARBOR_GIT_PROJECT"), os.Getenv("HARBOR_GIT_IMAGENAME"), os.Getenv("HARBOR_GIT_GITURI"), os.Getenv("HARBOR_GIT_STOREMETHOD"), os.Getenv("HARBOR_GIT_BRANCH"))
}
func initConfig(workspace, project, imagename, gituri, storemethod, branch string) {
	config.Workspace = workspace
	config.Project = project
	config.Imagename = imagename
	config.Gituri = gituri
	config.Storemethod = storemethod
	config.Branch = branch
}
