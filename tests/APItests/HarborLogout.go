// HarborLogout.go
package HarborApi

import (
	"net/http"
)

func (a HarborApi) HarborLogout() (int, error) {

	response, err := http.Get(a.basePath + "/logout")

	defer response.Body.Close()

	return response.StatusCode, err
}
