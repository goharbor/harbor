package aliacree

import "time"

const (
	endpointTpl = "cr.%s.aliyuncs.com"
)

type timeUnix int64

func (t timeUnix) ToTime() time.Time {
	return time.Unix(int64(t)/1000, 0)
}

func (t timeUnix) String() string {
	return t.ToTime().String()
}
