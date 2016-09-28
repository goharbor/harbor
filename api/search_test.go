package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/tests/apitests/apilib"
	"os/exec"
	"testing"
)

func TestSearch(t *testing.T) {
	fmt.Println("Testing Search(SearchGet) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()
	var result apilib.Search
	result, err := apiTest.SearchGet("library")
	//fmt.Printf("%+v\n", result)
	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(result.Projects[0].Id, int64(1), "Project id should be equal")
		assert.Equal(result.Projects[0].Name, "library", "Project name should be library")
		assert.Equal(result.Projects[0].Public, int32(1), "Project public status should be 1 (true)")
	}
	//case 2: push image and search
	command := `ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
	cmd := exec.Command("/bin/bash", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		t.Error("Error while push image ", err.Error())
		t.Log(err)
	}
	ip := string(out)
	ip = ip[0 : len(ip)-1]
	command1 := `docker pull registry:2.5.0`
	command2 := `docker tag registry:2.5.0 ` + ip + "/library/registry:2.5.0"
	command3 := `docker push ` + ip + `/library/registry:2.5.0`
	cmd = exec.Command("/bin/bash", "-c", command1, command2, command3)
	err = cmd.Run()
	if err != nil {
		t.Error("Error while push image ", err.Error())
		t.Log(err)
	}
	result, err = apiTest.SearchGet("registry")
	if err != nil {
		t.Error("Error while search project or repository", err.Error())
		t.Log(err)
	} else {
		assert.Equal(result.Repositories[0].ProjectId, int32(1), "Project id should be equal")
		assert.Equal(result.Repositories[0].ProjectName, "library", "Project name should be library")
		assert.Equal(result.Repositories[0].ProjectPublic, int32(1), "Project public status should be 1 (true)")
		assert.Equal(result.Repositories[0].RepositoryName, "registry", "Repository name should be registry")
	}

}
