package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vmware/harbor/git"
)

func InitClient() *git.Client {
	config := git.Pairs()
	client, err := git.NewClient(config.Workspace, config.Project, config.Imagename, config.Gituri, config.Branch)
	if err != nil {
		log.Error("the creation of git client failed")
	}
	err = client.Clone()
	if err != nil {
		log.Error(fmt.Sprintf("%s :%s", "git clone failed,git uri is", config.Gituri))
	}
	return client
}
func PullTimer(client *git.Client) {
	timer := time.NewTicker(time.Second * 2)
	for {
		<-timer.C
		if err := client.Pull(); err != nil {
			log.Error(fmt.Sprintf("%s:%s", "git pull failed,the uri is:", client.URI))
		}
	}
}
