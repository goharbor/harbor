package systemartifact

import (
	"context"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	registrytesting "github.com/goharbor/harbor/src/testing/pkg/registry"
	"github.com/goharbor/harbor/src/testing/pkg/systemartifact/cleanup"
	sysartifactdaotesting "github.com/goharbor/harbor/src/testing/pkg/systemartifact/dao"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

type ManagerTestSuite struct {
	suite.Suite
	regCli          *registrytesting.FakeClient
	dao             *sysartifactdaotesting.DAO
	mgr             *systemArtifactManager
	cleanupCriteria *cleanup.Selector
}

func (suite *ManagerTestSuite) SetupSuite() {

}

func (suite *ManagerTestSuite) SetupTest() {
	suite.regCli = &registrytesting.FakeClient{}
	suite.dao = &sysartifactdaotesting.DAO{}
	suite.cleanupCriteria = &cleanup.Selector{}
	suite.mgr = &systemArtifactManager{
		regCli:                  suite.regCli,
		dao:                     suite.dao,
		defaultCleanupCriterion: suite.cleanupCriteria,
		cleanupCriteria:         make(map[string]Selector),
	}
}

func (suite *ManagerTestSuite) TestCreate() {
	sa := model.SystemArtifact{
		Repository: "test_repo",
		Digest:     "test_digest",
		Size:       int64(100),
		Vendor:     "test_vendor",
		Type:       "test_type",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}
	suite.dao.On("Create", mock.Anything, &sa, mock.Anything).Return(int64(1), nil).Once()
	suite.regCli.On("PushBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	reader := strings.NewReader("test data string")
	id, err := suite.mgr.Create(orm.NewContext(nil, &ormtesting.FakeOrmer{}), &sa, reader)
	suite.Equalf(int64(1), id, "Expected row to correctly inserted")
	suite.NoErrorf(err, "Unexpected error when creating artifact: %v", err)
	suite.regCli.AssertCalled(suite.T(), "PushBlob")
}

func (suite *ManagerTestSuite) TestCreatePushBlobFails() {
	sa := model.SystemArtifact{
		Repository: "test_repo",
		Digest:     "test_digest",
		Size:       int64(100),
		Vendor:     "test_vendor",
		Type:       "test_type",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}
	suite.dao.On("Create", mock.Anything, &sa, mock.Anything).Return(int64(1), nil).Once()
	suite.dao.On("Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	suite.regCli.On("PushBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error")).Once()
	reader := strings.NewReader("test data string")
	id, err := suite.mgr.Create(orm.NewContext(nil, &ormtesting.FakeOrmer{}), &sa, reader)
	suite.Equalf(int64(0), id, "Expected no row to be inserted")
	suite.Errorf(err, "Expected error when creating artifact: %v", err)
	suite.dao.AssertCalled(suite.T(), "Create", mock.Anything, &sa, mock.Anything)
	suite.regCli.AssertCalled(suite.T(), "PushBlob")
}

func (suite *ManagerTestSuite) TestCreateArtifactRecordFailure() {
	sa := model.SystemArtifact{
		Repository: "test_repo",
		Digest:     "test_digest",
		Size:       int64(100),
		Vendor:     "test_vendor",
		Type:       "test_type",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}
	suite.dao.On("Create", mock.Anything, &sa, mock.Anything).Return(int64(0), errors.New("error")).Once()
	suite.regCli.On("PushBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	suite.regCli.On("PushBlob", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(nil).Once()

	reader := strings.NewReader("test data string")
	id, err := suite.mgr.Create(orm.NewContext(nil, &ormtesting.FakeOrmer{}), &sa, reader)
	suite.Equalf(int64(0), id, "Expected no row to be inserted")
	suite.Errorf(err, "Expected error when creating artifact: %v", err)
	suite.dao.AssertCalled(suite.T(), "Create", mock.Anything, mock.Anything)
	suite.regCli.AssertNotCalled(suite.T(), "PushBlob")
}

func (suite *ManagerTestSuite) TestRead() {
	sa := model.SystemArtifact{
		Repository: "test_repo",
		Digest:     "test_digest",
		Size:       int64(100),
		Vendor:     "test_vendor",
		Type:       "test_type",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	dummyRepoFilepath := fmt.Sprintf("/tmp/sys_art_test.dmp_%v", time.Now())
	data := []byte("test data")
	err := ioutil.WriteFile(dummyRepoFilepath, data, os.ModePerm)
	suite.NoErrorf(err, "Unexpected error when creating test repo file: %v", dummyRepoFilepath)

	repoHandle, err := os.Open(dummyRepoFilepath)
	suite.NoErrorf(err, "Unexpected error when reading test repo file: %v", dummyRepoFilepath)
	defer repoHandle.Close()

	suite.dao.On("Get", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(&sa, nil).Once()
	suite.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(len(data), repoHandle, nil).Once()

	readCloser, err := suite.mgr.Read(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.NoErrorf(err, "Unexpected error when reading artifact: %v", err)
	suite.dao.AssertCalled(suite.T(), "Get", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertCalled(suite.T(), "PullBlob")
	suite.NotNilf(readCloser, "Expected valid read closer instance but was nil")
}

func (suite *ManagerTestSuite) TestReadSystemArtifactRecordNotFound() {

	dummyRepoFilepath := fmt.Sprintf("/tmp/sys_art_test.dmp_%v", time.Now())
	data := []byte("test data")
	err := ioutil.WriteFile(dummyRepoFilepath, data, os.ModePerm)
	suite.NoErrorf(err, "Unexpected error when creating test repo file: %v", dummyRepoFilepath)

	repoHandle, err := os.Open(dummyRepoFilepath)
	suite.NoErrorf(err, "Unexpected error when reading test repo file: %v", dummyRepoFilepath)
	defer repoHandle.Close()

	errToRet := orm.ErrNoRows

	suite.dao.On("Get", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(nil, errToRet).Once()
	suite.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(len(data), repoHandle, nil).Once()

	readCloser, err := suite.mgr.Read(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.Errorf(err, "Expected error when reading artifact: %v", errToRet)
	suite.dao.AssertCalled(suite.T(), "Get", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertNotCalled(suite.T(), "PullBlob")
	suite.Nilf(readCloser, "Expected null read closer instance but was valid")
}

func (suite *ManagerTestSuite) TestDelete() {

	suite.dao.On("Delete", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(nil).Once()
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(nil).Once()

	err := suite.mgr.Delete(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.NoErrorf(err, "Unexpected error when deleting artifact: %v", err)
	suite.dao.AssertCalled(suite.T(), "Delete", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertCalled(suite.T(), "DeleteBlob")
}

func (suite *ManagerTestSuite) TestDeleteSystemArtifactDeleteError() {

	errToRet := orm.ErrNoRows
	suite.dao.On("Delete", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(errToRet).Once()
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(nil).Once()

	err := suite.mgr.Delete(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.Errorf(err, "Expected error when deleting artifact: %v", err)
	suite.dao.AssertCalled(suite.T(), "Delete", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertCalled(suite.T(), "DeleteBlob")
}

func (suite *ManagerTestSuite) TestDeleteSystemArtifactBlobDeleteError() {

	errToRet := orm.ErrNoRows
	suite.dao.On("Delete", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(nil).Once()
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(errToRet).Once()

	err := suite.mgr.Delete(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.Errorf(err, "Expected error when deleting artifact: %v", err)
	suite.dao.AssertNotCalled(suite.T(), "Delete", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertCalled(suite.T(), "DeleteBlob")
}

func (suite *ManagerTestSuite) TestExist() {
	sa := model.SystemArtifact{
		Repository: "test_repo",
		Digest:     "test_digest",
		Size:       int64(100),
		Vendor:     "test_vendor",
		Type:       "test_type",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	suite.dao.On("Get", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(&sa, nil).Once()
	suite.regCli.On("BlobExist", mock.Anything, mock.Anything).Return(true, nil).Once()

	exists, err := suite.mgr.Exists(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.NoErrorf(err, "Unexpected error when checking if artifact exists: %v", err)
	suite.dao.AssertCalled(suite.T(), "Get", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertCalled(suite.T(), "BlobExist")
	suite.True(exists, "Expected exists to be true but was false")
}

func (suite *ManagerTestSuite) TestExistSystemArtifactRecordReadError() {

	errToReturn := orm.ErrNoRows

	suite.dao.On("Get", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(nil, errToReturn).Once()
	suite.regCli.On("BlobExist", mock.Anything, mock.Anything).Return(true, nil).Once()

	exists, err := suite.mgr.Exists(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.Error(err, "Expected error when checking if artifact exists")
	suite.dao.AssertCalled(suite.T(), "Get", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertNotCalled(suite.T(), "BlobExist")
	suite.False(exists, "Expected exists to be false but was true")
}

func (suite *ManagerTestSuite) TestExistSystemArtifactBlobReadError() {

	sa := model.SystemArtifact{
		Repository: "test_repo",
		Digest:     "test_digest",
		Size:       int64(100),
		Vendor:     "test_vendor",
		Type:       "test_type",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	suite.dao.On("Get", mock.Anything, "test_vendor", "test_repo", "test_digest").Return(&sa, nil).Once()
	suite.regCli.On("BlobExist", mock.Anything, mock.Anything).Return(false, errors.New("test error")).Once()

	exists, err := suite.mgr.Exists(context.TODO(), "test_vendor", "test_repo", "test_digest")

	suite.Error(err, "Expected error when checking if artifact exists")
	suite.dao.AssertCalled(suite.T(), "Get", mock.Anything, "test_vendor", "test_repo", "test_digest")
	suite.regCli.AssertCalled(suite.T(), "BlobExist")
	suite.False(exists, "Expected exists to be false but was true")
}

func (suite *ManagerTestSuite) TestGetStorageSize() {

	suite.dao.On("Size", mock.Anything).Return(int64(400), nil).Once()

	size, err := suite.mgr.GetStorageSize(context.TODO())

	suite.NoErrorf(err, "Unexpected error encountered: %v", err)
	suite.dao.AssertCalled(suite.T(), "Size", mock.Anything)
	suite.Equalf(int64(400), size, "Expected size to be 400 but was : %v", size)
}

func (suite *ManagerTestSuite) TestGetStorageSizeError() {

	suite.dao.On("Size", mock.Anything).Return(int64(0), errors.New("test error")).Once()

	size, err := suite.mgr.GetStorageSize(context.TODO())

	suite.Errorf(err, "Expected error encountered: %v", err)
	suite.dao.AssertCalled(suite.T(), "Size", mock.Anything)
	suite.Equalf(int64(0), size, "Expected size to be 0 but was : %v", size)
}

func (suite *ManagerTestSuite) TestCleanupCriteriaRegistration() {
	vendor := "test_vendor"
	artifactType := "test_artifact_type"
	suite.mgr.RegisterCleanupCriteria(vendor, artifactType, suite)

	criteria := suite.mgr.GetCleanupCriteria(vendor, artifactType)
	suite.Equalf(suite, criteria, "Expected cleanup criteria to be the same as suite")

	criteria = suite.mgr.GetCleanupCriteria("test_vendor1", "test_artifact1")
	suite.Equalf(DefaultSelector, criteria, "Expected cleanup criteria to be the same as default cleanup criteria")
}

func (suite *ManagerTestSuite) TestCleanup() {
	sa1 := model.SystemArtifact{
		Repository: "test_repo1",
		Digest:     "test_digest1",
		Size:       int64(100),
		Vendor:     "test_vendor1",
		Type:       "test_type1",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	sa2 := model.SystemArtifact{
		Repository: "test_repo2",
		Digest:     "test_digest2",
		Size:       int64(300),
		Vendor:     "test_vendor2",
		Type:       "test_type2",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	sa3 := model.SystemArtifact{
		Repository: "test_repo3",
		Digest:     "test_digest3",
		Size:       int64(300),
		Vendor:     "test_vendor3",
		Type:       "test_type3",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	mockCleaupCriteria1 := cleanup.Selector{}
	mockCleaupCriteria1.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa1}, nil).Once()

	mockCleaupCriteria2 := cleanup.Selector{}
	mockCleaupCriteria2.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa2}, nil).Once()

	suite.cleanupCriteria.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa3}, nil).Once()

	suite.mgr.RegisterCleanupCriteria("test_vendor1", "test_type1", &mockCleaupCriteria1)
	suite.mgr.RegisterCleanupCriteria("test_vendor2", "test_type2", &mockCleaupCriteria2)

	suite.dao.On("Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(3)
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(nil).Times(3)

	totalDeleted, totalSizeReclaimed, err := suite.mgr.Cleanup(context.TODO())
	suite.Equalf(int64(3), totalDeleted, "System artifacts delete; Expected:%d, Actual:%d", int64(3), totalDeleted)
	suite.Equalf(int64(700), totalSizeReclaimed, "System artifacts delete; Expected:%d, Actual:%d", int64(700), totalDeleted)
	suite.NoErrorf(err, "Unexpected error: %v", err)
}

func (suite *ManagerTestSuite) TestCleanupError() {
	sa1 := model.SystemArtifact{
		Repository: "test_repo13000",
		Digest:     "test_digest13000",
		Size:       int64(100),
		Vendor:     "test_vendor13000",
		Type:       "test_type13000",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	sa3 := model.SystemArtifact{
		Repository: "test_repo33000",
		Digest:     "test_digest33000",
		Size:       int64(300),
		Vendor:     "test_vendor33000",
		Type:       "test_type33000",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	mockCleaupCriteria1 := cleanup.Selector{}
	mockCleaupCriteria1.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa1}, nil).Once()

	mockCleaupCriteria2 := cleanup.Selector{}
	mockCleaupCriteria2.On("List", mock.Anything).Return(nil, errors.New("test error")).Once()

	suite.cleanupCriteria.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa3}, nil)

	suite.mgr.RegisterCleanupCriteria("test_vendor13000", "test_type13000", &mockCleaupCriteria1)
	suite.mgr.RegisterCleanupCriteria("test_vendor23000", "test_type23000", &mockCleaupCriteria2)

	suite.dao.On("Delete", mock.Anything, "test_vendor13000", "test_repo13000", "test_digest13000").Return(nil)
	suite.dao.On("Delete", mock.Anything, "test_vendor33000", "test_repo33000", "test_digest33000").Return(nil)
	suite.dao.On("Delete", mock.Anything, "test_vendor23000", "test_repo23000", mock.Anything).Return(nil)
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(nil)

	totalDeleted, totalSizeReclaimed, err := suite.mgr.Cleanup(context.TODO())
	suite.Equalf(int64(2), totalDeleted, "System artifacts delete; Expected:%d, Actual:%d", int64(2), totalDeleted)
	suite.Equalf(int64(400), totalSizeReclaimed, "System artifacts delete; Expected:%d, Actual:%d", int64(400), totalDeleted)
	suite.NoError(err, "Expected no error but was %v", err)
}

func (suite *ManagerTestSuite) TestCleanupErrorDefaultCriteria() {
	sa1 := model.SystemArtifact{
		Repository: "test_repo1",
		Digest:     "test_digest1",
		Size:       int64(100),
		Vendor:     "test_vendor1",
		Type:       "test_type1",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	sa2 := model.SystemArtifact{
		Repository: "test_repo2",
		Digest:     "test_digest2",
		Size:       int64(300),
		Vendor:     "test_vendor2",
		Type:       "test_type2",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	mockCleaupCriteria1 := cleanup.Selector{}
	mockCleaupCriteria1.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa1}, nil).Once()

	mockCleaupCriteria2 := cleanup.Selector{}
	mockCleaupCriteria2.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa2}, nil).Once()

	suite.cleanupCriteria.On("List", mock.Anything).Return(nil, errors.New("test error"))

	suite.mgr.RegisterCleanupCriteria("test_vendor1", "test_type1", &mockCleaupCriteria1)
	suite.mgr.RegisterCleanupCriteria("test_vendor2", "test_type2", &mockCleaupCriteria2)

	suite.dao.On("Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(nil)

	totalDeleted, totalSizeReclaimed, err := suite.mgr.Cleanup(context.TODO())
	suite.Equalf(int64(2), totalDeleted, "System artifacts delete; Expected:%d, Actual:%d", int64(2), totalDeleted)
	suite.Equalf(int64(400), totalSizeReclaimed, "System artifacts delete; Expected:%d, Actual:%d", int64(400), totalDeleted)
	suite.NoErrorf(err, "Expected no error but was %v", err)
}

func (suite *ManagerTestSuite) TestCleanupErrorForVendor() {
	sa1 := model.SystemArtifact{
		Repository: "test_repo10000",
		Digest:     "test_digest10000",
		Size:       int64(100),
		Vendor:     "test_vendor10000",
		Type:       "test_type10000",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	sa2 := model.SystemArtifact{
		Repository: "test_repo20000",
		Digest:     "test_digest20000",
		Size:       int64(300),
		Vendor:     "test_vendor10000",
		Type:       "test_type10000",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	sa3 := model.SystemArtifact{
		Repository: "test_repo30000",
		Digest:     "test_digest30000",
		Size:       int64(300),
		Vendor:     "test_vendor30000",
		Type:       "test_type30000",
		CreateTime: time.Now(),
		ExtraAttrs: "",
	}

	mockCleaupCriteria1 := cleanup.Selector{}
	mockCleaupCriteria1.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa1, &sa2}, nil).Times(2)

	suite.cleanupCriteria.On("List", mock.Anything).Return([]*model.SystemArtifact{&sa3}, nil).Times(2)

	suite.mgr.RegisterCleanupCriteria("test_vendor10000", "test_type10000", &mockCleaupCriteria1)

	suite.dao.On("Delete", mock.Anything, "test_vendor10000", "test_repo10000", "test_digest10000").Return(nil).Once()
	suite.dao.On("Delete", mock.Anything, "test_vendor10000", "test_repo20000", "test_digest20000").Return(errors.New("test error")).Once()
	suite.dao.On("Delete", mock.Anything, "test_vendor30000", "test_repo30000", "test_digest30000").Return(nil).Once()
	suite.regCli.On("DeleteBlob", mock.Anything, mock.Anything).Return(nil).Times(3)

	totalDeleted, totalSizeReclaimed, err := suite.mgr.Cleanup(context.TODO())
	suite.Equalf(int64(2), totalDeleted, "System artifacts delete; Expected:%d, Actual:%d", int64(2), totalDeleted)
	suite.Equalf(int64(400), totalSizeReclaimed, "System artifacts delete; Expected:%d, Actual:%d", int64(400), totalDeleted)
	suite.NoErrorf(err, "Expected no error, but was %v", err)
}

func (suite *ManagerTestSuite) List(ctx context.Context) ([]*model.SystemArtifact, error) {
	return make([]*model.SystemArtifact, 0), nil
}

func (suite *ManagerTestSuite) ListWithFilters(ctx context.Context, query *q.Query) ([]*model.SystemArtifact, error) {
	return make([]*model.SystemArtifact, 0), nil
}

func TestManagerTestSuite(t *testing.T) {
	mgr := &ManagerTestSuite{}
	suite.Run(t, mgr)
}
