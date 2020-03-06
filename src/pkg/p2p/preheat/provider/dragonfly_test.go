package provider

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
)

var dfDriver = &DragonflyDriver{
	instance: &models.Metadata{
		Endpoint: "http://47.111.250.75:8002",
		AuthMode: auth.AuthModeNone,
	},
}

// Test for @DragonflyDriver.GetHealth.
func TestDragonflyGetHealth(t *testing.T) {
	status, err := dfDriver.GetHealth()
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != "Healthy" {
		t.Errorf("expect running but got %s", status.Status)
	}
}

// Test for @DragonflyDriver.Preheat and @DragonflyDriver.CheckProgress.
func TestDragonflyPreheat(t *testing.T) {
	image := &PreheatImage{
		Type: "image",
		URL:  "https://registry.cn-hangzhou.aliyuncs.com/v2/acs/alpine/manifests/latest",
	}
	status, err := dfDriver.Preheat(image)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "preheat task already exists") {
			index := strings.Index(errMsg, "id:")
			if index > 0 {
				id := strings.Trim(errMsg[index+3:], ` :}"`)
				status = &PreheatingStatus{TaskID: id, Status: "PENDING"}
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	}
	time.Sleep(time.Second)
	for status.Status == "PENDING" || status.Status == "RUNNING" {
		status, err = dfDriver.CheckProgress(status.TaskID)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(status)
		time.Sleep(time.Second)
	}
	if status.Status == "FAILED" {
		t.Errorf("expect success but got %v", status)
	}
}
