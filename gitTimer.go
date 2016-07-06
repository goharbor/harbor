package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vmware/harbor/service"
)

func InitClient() *service.Client {
	config := service.Pairs()
	client, err := service.NewClient(config.Workspace, config.Project, config.Imagename, config.Gituri, config.Branch)
	if err != nil {
		log.Error("the creation of git client failed")
	}
	err = service.Clone()
	if err != nil {
		log.Error(fmt.Sprintf("%s :%s", "git clone failed,git uri is", config.Gituri))
	}
	return &client
}
func PullTimer(client *service.Client) {
	timer := time.NewTicker(time.Second * 2)
	for {
		<-timer.C
		if err := service.Pull(); err != nil {
			log.Error(fmt.Sprintf("%s:%s", "git pull failed,the uri is:", client.URI))
		}
	}
}
