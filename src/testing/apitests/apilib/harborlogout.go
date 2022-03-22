// HarborLogout.go
package apilib

import (
	"net/http"
)

func (a HarborAPI) HarborLogout() (int, error) {
	response, err := http.Get(a.basePath + "/logout")
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	return response.StatusCode, nil
}
