package handlers

import (
	"reflect"
	"testing"

	"github.com/Sirupsen/logrus"
	ctxu "github.com/docker/distribution/context"
	"github.com/docker/notary"
	"github.com/docker/notary/server/storage"
	"github.com/stretchr/testify/require"
)

type changefeedArgs struct {
	logger   ctxu.Logger
	store    storage.MetaStore
	gun      string
	changeID string
	pageSize int64
}

type changefeedTest struct {
	name    string
	args    changefeedArgs
	want    []byte
	wantErr bool
}

func Test_changefeed(t *testing.T) {
	s := storage.NewMemStorage()

	tests := []changefeedTest{
		{
			name: "Empty Store",
			args: changefeedArgs{
				logger:   logrus.New(),
				store:    s,
				gun:      "",
				changeID: "0",
				pageSize: notary.DefaultPageSize,
			},
			want:    []byte("{\"count\":0,\"records\":null}"),
			wantErr: false,
		},
		{
			name: "Bad ChangeID",
			args: changefeedArgs{
				logger:   logrus.New(),
				store:    s,
				gun:      "",
				changeID: "not_a_number",
				pageSize: notary.DefaultPageSize,
			},
			want:    nil,
			wantErr: true,
		},
	}
	runChangefeedTests(t, tests)

}

func runChangefeedTests(t *testing.T, tests []changefeedTest) {
	for _, tt := range tests {
		got, err := changefeed(tt.args.logger, tt.args.store, tt.args.gun, tt.args.changeID, tt.args.pageSize)
		if tt.wantErr {
			require.Error(t, err,
				"%q. changefeed() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		require.True(t, reflect.DeepEqual(got, tt.want),
			"%q. changefeed() = %v, want %v", tt.name, string(got), string(tt.want))
	}
}

func Test_checkChangefeedInputs(t *testing.T) {
	type args struct {
		logger ctxu.Logger
		s      interface{}
		ps     string
	}
	s := storage.NewMemStorage()
	tests := []struct {
		name         string
		args         args
		wantStore    storage.MetaStore
		wantPageSize int64
		wantErr      bool
	}{
		// Error cases
		{
			name: "No MetaStore",
			args: args{
				logger: logrus.New(),
				s:      nil,
			},
			wantErr: true,
		},
		{
			name: "Bad page size",
			args: args{
				logger: logrus.New(),
				s:      s,
				ps:     "not_a_number",
			},
			wantErr:   true,
			wantStore: s,
		},
		{
			name: "Zero page size",
			args: args{
				logger: logrus.New(),
				s:      s,
				ps:     "0",
			},
			wantStore:    s,
			wantPageSize: notary.DefaultPageSize,
		},
		{
			name: "Non-zero Page Size",
			args: args{
				logger: logrus.New(),
				s:      s,
				ps:     "10",
			},
			wantStore:    s,
			wantPageSize: 10,
		},
		{
			name: "Reversed \"false\"",
			args: args{
				logger: logrus.New(),
				s:      s,
				ps:     "-10",
			},
			wantStore:    s,
			wantPageSize: -10,
		},
	}
	for _, tt := range tests {
		gotStore, gotPageSize, err := checkChangefeedInputs(tt.args.logger, tt.args.s, tt.args.ps)
		if tt.wantErr {
			require.Error(t, err,
				"%q. checkChangefeedInputs() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		require.True(t, reflect.DeepEqual(gotStore, tt.wantStore),
			"%q. checkChangefeedInputs() gotStore = %v, want %v", tt.name, gotStore, tt.wantStore)

		require.Equal(t, tt.wantPageSize, gotPageSize,
			"%q. checkChangefeedInputs() gotPageSize = %v, want %v", tt.name, gotPageSize, tt.wantPageSize)

	}
}
