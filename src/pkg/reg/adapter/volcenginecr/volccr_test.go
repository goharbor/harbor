package volcenginecr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolccr_createNamespace(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	err := a.createNamespace("ut_test")
	assert.Error(t, err)
}

func TestVolccr_createRepository(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	err := a.createRepository("ut_test", "ut_test")
	assert.Error(t, err)
}

func TestVolccr_deleteTags(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	err := a.deleteTags("ut_test", "ut_test", []*string{})
	assert.Error(t, err)
}

func TestVolccr_listCandidateNamespaces(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	_, err := a.listCandidateNamespaces("ut_test")
	assert.Error(t, err)
}

func TestVolccr_listNamespaces(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	_, err := a.listNamespaces()
	assert.Error(t, err)
}

func TestVolccr_listRepositories(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	_, err := a.listRepositories("ut_test")
	assert.Error(t, err)
}

func TestVolccr_listAllTags(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	_, err := a.listAllTags("ut_test", "ut_test")
	assert.Error(t, err)
}

func TestVolccr_listCandidateTags(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	_, err := a.listCandidateTags("ut_test", "ut_test", "sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc")
	assert.Error(t, err)
}

func TestVolccr_namespaceExist(t *testing.T) {
	a, s := getMockAdapter_withoutCred(t, true, true)
	defer s.Close()
	_, err := a.namespaceExist("ut_test")
	assert.Error(t, err)
}
