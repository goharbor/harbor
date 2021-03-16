package exporter

import (
	"net/http"
	"testing"
)

func TestHarborClient_harborURL(t *testing.T) {
	type fields struct {
		HarborScheme string
		HarborHost   string
		HarborPort   int
		Client       *http.Client
	}
	type args struct {
		p string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "Can get URL", fields: fields{
			HarborScheme: "https",
			HarborHost:   "test.harbor.dev",
			HarborPort:   8443,
			Client:       http.DefaultClient,
		},
			args: args{p: ""}, want: "https://test.harbor.dev:8443",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := HarborClient{
				HarborScheme: tt.fields.HarborScheme,
				HarborHost:   tt.fields.HarborHost,
				HarborPort:   tt.fields.HarborPort,
				Client:       tt.fields.Client,
			}
			if got := hc.harborURL(tt.args.p); got.String() != tt.want {
				t.Errorf("HarborClient.harborURL() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}
