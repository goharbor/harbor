// HarborLogon.go
package apilib

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func (a HarborAPI) HarborLogin(user UsrInfo) (int, error) {

	v := url.Values{}
	v.Set("principal", user.Name)
	v.Set("password", user.Passwd)

	body := ioutil.NopCloser(strings.NewReader(v.Encode())) // endode v:[body struce]

	client := &http.Client{}
	reqest, err := http.NewRequest("POST", a.basePath+"/login", body)
	if err != nil {
		return 0, err
	}
	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded;param=value") // setting post head

	resp, err := client.Do(reqest)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close() // close resp.Body

	return resp.StatusCode, nil
}
