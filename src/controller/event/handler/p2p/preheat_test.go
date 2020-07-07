package p2p

import (
	"context"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/p2p/preheat"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/q"
	pkg_artifact "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type fakeController struct {
}

func (fc *fakeController) Ensure(ctx context.Context, repository, digest string, tags ...string) (bool, int64, error) {
	return true, 10, nil
}

func (fc *fakeController) Count(ctx context.Context, query *q.Query) (int64, error) {
	return 0, nil
}

func (fc *fakeController) List(ctx context.Context, query *q.Query, option *artifact.Option) ([]*artifact.Artifact, error) {
	return nil, nil
}

func (fc *fakeController) Get(ctx context.Context, id int64, option *artifact.Option) (*artifact.Artifact, error) {
	return &artifact.Artifact{
		Artifact: pkg_artifact.Artifact{
			Type: "IMAGE",
		},
	}, nil
}

func (fc *fakeController) GetByReference(ctx context.Context, repository, reference string, option *artifact.Option) (*artifact.Artifact, error) {
	return &artifact.Artifact{}, nil
}

func (fc *fakeController) Delete(ctx context.Context, id int64) error {
	return nil
}

func (fc *fakeController) Copy(ctx context.Context, srcRepo, reference, dstRepo string) (int64, error) {
	return 0, nil
}

func (fc *fakeController) UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) error {
	return nil
}

func (fc *fakeController) GetAddition(ctx context.Context, artifactID int64, addition string) (*processor.Addition, error) {
	return nil, nil
}

func (fc *fakeController) AddLabel(ctx context.Context, artifactID int64, labelID int64) error {
	return nil
}

func (fc *fakeController) RemoveLabel(ctx context.Context, artifactID int64, labelID int64) error {
	return nil
}

func (fc *fakeController) Walk(ctx context.Context, root *artifact.Artifact, walkFn func(*artifact.Artifact) error, option *artifact.Option) error {
	return nil
}

type fakeEnforcer struct{}

// EnforcePolicy enforces preheating action by the given policy
func (fe *fakeEnforcer) EnforcePolicy(ctx context.Context, policyID int64) (int64, error) {
	return 0, nil
}

// PreheatArtifact enforces preheating action by the given artifact.
func (fe *fakeEnforcer) PreheatArtifact(ctx context.Context, art *artifact.Artifact) ([]int64, error) {
	return nil, nil
}

// TestPreheatHandler_Handle ...
func TestPreheatHandler_Handle(t *testing.T) {
	config.Init()

	ArtifactCtl := artifact.Ctl
	PreheatEnf := preheat.Enf
	defer func() {
		artifact.Ctl = ArtifactCtl
		preheat.Enf = PreheatEnf
	}()
	artifact.Ctl = &fakeController{}
	preheat.Enf = &fakeEnforcer{}

	handler := &Handler{
		Context: context.Background,
	}

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PreheatHandler String Error",
			args: args{
				data: "",
			},
			wantErr: true,
		},
		{
			name: "PreheatHandler 1",
			args: args{
				data: &event.PushArtifactEvent{
					ArtifactEvent: &event.ArtifactEvent{
						Artifact: &pkg_artifact.Artifact{
							Type:         "IMAGE",
							ID:           11,
							RepositoryID: 23,
						},
						Tags: []string{"v1.1", "v1.2"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "PreheatHandler 2",
			args: args{
				data: &event.ScanImageEvent{
					OccurAt: time.Now(),
					Artifact: &v1.Artifact{
						Repository: "library",
						Digest:     "sha256:1359608115b94599e5641638bac5aef1ddfaa79bb96057ebf41ebc8d33acf8a7",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "PreheatHandler 3",
			args: args{
				data: &event.ArtifactLabeledEvent{
					OccurAt:    time.Now(),
					ArtifactID: 1,
					LabelID:    2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestPreheatHandler_IsStateful(t *testing.T) {
	handler := &Handler{}
	assert.False(t, handler.IsStateful())
}
