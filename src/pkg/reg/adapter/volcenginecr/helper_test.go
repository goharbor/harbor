package volcenginecr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getRegionRegistryNamer(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantRegion   string
		wantRegistry string
		wantErr      bool
	}{
		{"registry beijing", "https://enterprise-cn-beijing.cr.volces.com", "cn-beijing", "enterprise", false},
		{"invalid url", "http://enterprise-cn-beijing.cr.volces.com", "", "", true},
		{"invalid region", "https://enterprise-us-test.cr.volces.com", "", "", true},
		{"invalid suffix", "https://enterprise-us-test.cr-test.volces.com", "", "", true},
		{"registry shanghai", "https://cn-beijing-cn-shanghai.cr.volces.com", "cn-shanghai", "cn-beijing", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRegion, gotRegistry, err := getRegionRegistryName(tt.url)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tt.wantRegion, gotRegion)
			assert.Equal(t, tt.wantRegistry, gotRegistry)
		})
	}
}

func Test_getRealmService(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		insecure bool
		wantErr  bool
	}{
		{"ping success", "https://cr-cn-beijing.volces.com", false, false},
		{"ping success", "https://cr-cn-beijing.volces.com", true, false},
		{"ping error", "https://cr-test-cn-beijing.volces.com", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := getRealmService(tt.host, tt.insecure)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
