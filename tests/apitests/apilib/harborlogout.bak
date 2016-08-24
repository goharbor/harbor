// HarborLogout.go
package HarborAPI

import (
	"net/http"
)

func (a HarborAPI) HarborLogout() (int, error) {

	response, err := http.Get(a.basePath + "/logout")

	defer response.Body.Close()

	return response.StatusCode, err
}
