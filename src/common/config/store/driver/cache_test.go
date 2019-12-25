package driver

import (
	"testing"
	"time"

	bc "github.com/astaxie/beego/cache"
	"github.com/goharbor/harbor/src/common"
	"github.com/stretchr/testify/assert"
)

func TestCachedDriver_updateInterval(t *testing.T) {
	type fields struct {
		cache    bc.Cache
		Interval time.Duration
		Driver   Driver
	}
	type args struct {
		cfg map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		expect time.Duration
	}{
		{name: "zero", fields: fields{Interval: 0}, args: args{cfg: map[string]interface{}{common.CfgCacheIntervalSeconds: 0}}, expect: 0},
		{name: "normal", fields: fields{Interval: 0}, args: args{cfg: map[string]interface{}{common.CfgCacheIntervalSeconds: 10}}, expect: 10 * time.Second},
		{name: "negative", fields: fields{Interval: 0}, args: args{cfg: map[string]interface{}{common.CfgCacheIntervalSeconds: -10}}, expect: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CachedDriver{
				cache:    tt.fields.cache,
				Interval: tt.fields.Interval,
				Driver:   tt.fields.Driver,
			}
			c.updateInterval(tt.args.cfg)
			assert.Equal(t, tt.expect, c.Interval)
		})
	}
}
