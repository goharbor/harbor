package HarborAPI

import ()

type Search struct {
	Projects     []Project4Search    `json:"project,omitempty"`
	Repositories []Repository4Search `json:"repository,omitempty"`
}
