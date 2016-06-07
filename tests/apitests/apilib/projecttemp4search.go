package HarborAPI

import ()

type Project4Search struct {
	ProjectId    int32  `json:"id,omitempty"`
	ProjectName  string `json:"name,omitempty"`
	Public       int32   `json:"public,omitempty"`
}
