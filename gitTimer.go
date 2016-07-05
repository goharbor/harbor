package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/harbor/service"
)

func initClient() *service.Client {
	client, err := git.NewClient(os.Getenv("workspace"), os.Getenv("project"), os.Getenv("imageName"), os.Getenv("gitUri"), os.Getenv("branchName"))
	if err != nil {
		log.Error("the creation of git client failed")
	}
	err = client.Clone()
	if err != nil {
		log.Error(fmt.Sprintf("%s :%s", "git clone failed,git uri is", os.Getenv("gitUri")))
	}
	return &client
}
func pullTimer(client *service.Client) {
	timer := time.NewTicker(time.Second * 2)
	for {
		<-timer.C
		if err := client.Pull(); err != nil {
			log.Error(fmt.Sprintf("%s:%s", "git pull failed,the uri is:", client.URI))
		}
	}
}
