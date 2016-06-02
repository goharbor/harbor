package HarborApi

import ()

type Search struct {
	Projects     []Project    `json:"project,omitempty"`
	Repositories []Repository `json:"repositorie,omitempty"`
}
