package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVolumeInfo(t *testing.T) {
	fmt.Println("Testing Get Volume Info")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//case 1: get volume info without admin role
	CommonAddUser()
	code, _, err := apiTest.VolumeInfoGet(*testUser)
	if err != nil {
		t.Error("Error occured while get system volume info")
		t.Log(err)
	} else {
		assert.Equal(403, code, "Get system volume info should be 403")
	}
	//case 2: get volume info with admin role
	code, info, err := apiTest.VolumeInfoGet(*admin)
	if err != nil {
		t.Error("Error occured while get system volume info")
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get system volume info should be 200")
		if info.HarborStorage.Total <= 0 {
			assert.Equal(1, info.HarborStorage.Total, "Total storage of system should be larger than 0")
		}
		if info.HarborStorage.Free <= 0 {
			assert.Equal(1, info.HarborStorage.Free, "Free storage of system should be larger than 0")
		}
	}

}

func TestGetCert(t *testing.T) {
	fmt.Println("Testing Get Cert")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//case 1: get cert without admin role
	code, _, err := apiTest.CertGet(*testUser)
	if err != nil {
		t.Error("Error occured while get system cert")
		t.Log(err)
	} else {
		assert.Equal(403, code, "Get system cert should be 403")
	}
	//case 2: get cert with admin role
	code, content, err := apiTest.CertGet(*admin)
	if err != nil {
		t.Error("Error occured while get system cert")
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get system cert should be 200")
		assert.Equal("test for ca.crt.\n", string(content), "Get system cert content should be equal")

	}
	CommonDelUser()
}
