package huawei

import (
	"os"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/replication/model"
)

var HWAdapter adapter

func init() {
	hwRegistry := &model.Registry{
		ID:          1,
		Name:        "Huawei",
		Description: "Adapter for SWR -- The image registry of Huawei Cloud",
		Type:        model.RegistryTypeHuawei,
		URL:         "https://swr.cn-north-1.myhuaweicloud.com",
		Credential:  &model.Credential{AccessKey: "cn-north-1@IJYZLFBKBFN8LOUITAH", AccessSecret: "f31e8e2b948265afdae32e83722a7705fd43e154585ff69e64108247750e5d"},
		Insecure:    false,
		Status:      "",
	}
	adp, err := newAdapter(hwRegistry)
	if err != nil {
		os.Exit(1)
	}
	HWAdapter = *adp.(*adapter)
}

func TestAdapter_FetchImages(t *testing.T) {
	resources, err := HWAdapter.FetchImages(nil)
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

func TestAdapter_ManifestExist(t *testing.T) {
	exist, digest, err := HWAdapter.ManifestExist("", "")
	if err != nil {
		if strings.HasPrefix(err.Error(), "[401]") {
			t.Log("huawei ak/sk is not available", err.Error())
		} else {
			t.Error(err)
		}
	} else {
		if exist {
			t.Log(digest)
		}
	}
}

func TestAdapter_DeleteManifest(t *testing.T) {
	err := HWAdapter.DeleteManifest("sundaymango_mango/hello-world", "latest")
	if err != nil {
		if strings.HasPrefix(err.Error(), "[401]") {
			t.Log("huawei ak/sk is not available", err.Error())
		} else {
			t.Error(err)
		}
	} else {
		t.Error("the manifest is deleted")
	}
}
