package handler

import (
	"reflect"
	"testing"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

func Test_convertProvidersToFrontend(t *testing.T) {
	backend, _ := provider.ListProviders()
	tests := []struct {
		name         string
		backend      []*provider.Metadata
		wantFrontend []*models.Metadata
	}{
		{"",
			backend,
			[]*models.Metadata{
				{ID: "dragonfly", Icon: "https://raw.githubusercontent.com/alibaba/Dragonfly/master/docs/images/logo.png", Maintainers: []string{"Jin Zhang/taiyun.zj@alibaba-inc.com"}, Name: "Dragonfly", Source: "https://github.com/alibaba/Dragonfly", Version: "0.10.1"},
				{Icon: "https://github.com/uber/kraken/blob/master/assets/kraken-logo-color.svg", ID: "kraken", Maintainers: []string{"mmpei/peimingming@corp.netease.com"}, Name: "Kraken", Source: "https://github.com/uber/kraken", Version: "0.1.3"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFrontend := convertProvidersToFrontend(tt.backend); !reflect.DeepEqual(gotFrontend, tt.wantFrontend) {
				t.Errorf("convertProvidersToFrontend() = %#v, want %#v", gotFrontend, tt.wantFrontend)
			}
		})
	}
}
