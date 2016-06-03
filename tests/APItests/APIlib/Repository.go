package HarborAPI

import (
	"time"
)

type Repository struct {
	Id            string    `json:"id,omitempty"`
	Parent        string    `json:"parent,omitempty"`
	Created       time.Time `json:"created,omitempty"`
	DurationDays  string    `json:"duration_days,omitempty"`
	Author        string    `json:"author,omitempty"`
	Architecture  string    `json:"architecture,omitempty"`
	DockerVersion string    `json:"docker_version,omitempty"`
	Os            string    `json:"os,omitempty"`
}
