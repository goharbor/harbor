package huawei

import (
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/replication/model"
)

var HWAdapter Adapter

func init() {
	hwRegistry := &model.Registry{
		ID:          1,
		Name:        "Huawei",
		Description: "Adapter for SWR -- The image registry of Huawei Cloud",
		Type:        "huawei",
		URL:         "https://swr.cn-north-1.myhuaweicloud.com",
		Credential:  &model.Credential{AccessKey: "cn-north-1@AQR6NF5G2MQ1V7U4FCD", AccessSecret: "2f7ec95070592fd4838a3aa4fd09338c047fd1cd654b3422197318f97281cd9"},
		Insecure:    false,
		Status:      "",
	}
	HWAdapter.Registry = hwRegistry
}

func TestAdapter_FetchImages(t *testing.T) {
	resources, err := HWAdapter.FetchImages([]string{"swr_namespace2", "sunday0615"}, nil)
	if err != nil {
		if strings.HasPrefix(err.Error(), "[401]") {
			t.Log("huawei ak/sk is not available", err.Error())
		} else {
			t.Error(err)
		}
	} else {
		for _, resource := range resources {
			t.Log(*resource)
		}
	}
}
