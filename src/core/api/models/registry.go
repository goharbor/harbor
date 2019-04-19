package models

// RegistryUpdateRequest is request used to update a registry.
type RegistryUpdateRequest struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	URL            *string `json:"url"`
	CredentialType *string `json:"credential_type"`
	AccessKey      *string `json:"access_key"`
	AccessSecret   *string `json:"access_secret"`
	Insecure       *bool   `json:"insecure"`
}
