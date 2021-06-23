package util

import (
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common"
)

func TestMain(m *testing.M) {
	// do some initialization
	test.InitDatabaseFromEnv()
	os.Exit(m.Run())
}

func TestBuildImageResourceURL(t *testing.T) {
	cfg := map[string]interface{}{
		common.ExtEndpoint: "https://demo.goharbor.io",
	}
	config.InitWithSettings(cfg)

	type args struct {
		repoName  string
		reference string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"digest",
			args{"library/photon", "sha256:1f240846ba3fc84aaa3c5eeb24e6b119394e9c3bf9536371d485a8a405e3deb3"},
			"demo.goharbor.io/library/photon@sha256:1f240846ba3fc84aaa3c5eeb24e6b119394e9c3bf9536371d485a8a405e3deb3",
			false,
		},
		{
			"tag",
			args{"library/photon", "2.0"},
			"demo.goharbor.io/library/photon:2.0",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildImageResourceURL(tt.args.repoName, tt.args.reference)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildImageResourceURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildImageResourceURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
