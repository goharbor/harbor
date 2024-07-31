package volcenginecr

import (
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArtifactRegistry_DeleteManifest(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	err := a.DeleteManifest("ut_test/ut_test", "sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc")
	assert.Error(t, err)
}

func TestArtifactRegistry_DeleteTag(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	err := a.DeleteTag("ut_test/ut_test", "v1")
	assert.Error(t, err)
}

func TestArtifactRegistry_FetchArtifacts(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	tests := []struct {
		name    string
		filter  model.Filter
		wantErr bool
	}{
		{"filter name",
			model.Filter{
				Type:  model.FilterTypeName,
				Value: "ut_test",
			},
			true},
		{"filter tag",
			model.Filter{
				Type:  model.FilterTypeTag,
				Value: "v1",
			},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := a.FetchArtifacts([]*model.Filter{&tt.filter})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
