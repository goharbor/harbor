package exporter

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
)

type SysCollectorSuite struct {
	suite.Suite
}

func (c *SysCollectorSuite) SetupTest() {
	CacheInit(&Opt{
		CacheDuration: 1,
	})
}

func TestNewSystemInfoCollector(t *testing.T) {
	type args struct {
		hbrCli *HarborClient
	}
	tests := []struct {
		name string
		args args
		want *SystemInfoCollector
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSystemInfoCollector(tt.args.hbrCli); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSystemInfoCollector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemInfoCollector_Describe(t *testing.T) {
	type fields struct {
		HarborClient *HarborClient
	}
	type args struct {
		c chan<- *prometheus.Desc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &SystemInfoCollector{
				HarborClient: tt.fields.HarborClient,
			}
			hc.Describe(tt.args.c)
		})
	}
}

func TestSystemInfoCollector_Collect(t *testing.T) {
	type fields struct {
		HarborClient *HarborClient
	}
	type args struct {
		c chan<- prometheus.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &SystemInfoCollector{
				HarborClient: tt.fields.HarborClient,
			}
			hc.Collect(tt.args.c)
		})
	}
}

func TestSystemInfoCollector_getSysInfo(t *testing.T) {
	type fields struct {
		HarborClient *HarborClient
	}
	tests := []struct {
		name   string
		fields fields
		want   []prometheus.Metric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &SystemInfoCollector{
				HarborClient: tt.fields.HarborClient,
			}
			if got := hc.getSysInfo(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SystemInfoCollector.getSysInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
