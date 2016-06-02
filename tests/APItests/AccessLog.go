package HarborApi

type AccessLog struct {
	Username       string `json:"username,omitempty"`
	Keywords       string `json:"keywords,omitempty"`
	BeginTimestamp int32  `json:"beginTimestamp,omitempty"`
	EndTimestamp   int32  `json:"endTimestamp,omitempty"`
}
