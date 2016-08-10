package apilib

//Project4Search ...
type Project4Search struct {
	ProjectID    int32  `json:"id,omitempty"`
	ProjectName  string `json:"name,omitempty"`
	Public       int32   `json:"public,omitempty"`
}
