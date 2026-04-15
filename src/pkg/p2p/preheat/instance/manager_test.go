package instance

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/q"
	dao "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/instance"
	providerModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"

	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

type fakeDao struct {
	mock.Mock
}

var _ dao.DAO = (*fakeDao)(nil)
var lists = []*providerModel.Instance{
	{Name: "abc"},
}

func (d *fakeDao) Create(ctx context.Context, instance *providerModel.Instance) (int64, error) {
	var args = d.Called()
	return int64(args.Int(0)), args.Error(1)
}

func (d *fakeDao) Get(ctx context.Context, id int64) (*providerModel.Instance, error) {
	var args = d.Called()
	var instance *providerModel.Instance
	if args.Get(0) != nil {
		instance = args.Get(0).(*providerModel.Instance)
	}
	return instance, args.Error(1)
}

func (d *fakeDao) GetByName(ctx context.Context, name string) (*providerModel.Instance, error) {
	var args = d.Called()
	var instance *providerModel.Instance
	if args.Get(0) != nil {
		instance = args.Get(0).(*providerModel.Instance)
	}
	return instance, args.Error(1)
}

func (d *fakeDao) Update(ctx context.Context, instance *providerModel.Instance, props ...string) error {
	var args = d.Called()
	return args.Error(0)
}

func (d *fakeDao) Delete(ctx context.Context, id int64) error {
	var args = d.Called()
	return args.Error(0)
}

func (d *fakeDao) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	var args = d.Called()

	return int64(args.Int(0)), args.Error(1)
}

func (d *fakeDao) List(ctx context.Context, query *q.Query) (ins []*providerModel.Instance, err error) {
	var args = d.Called()
	var instances []*providerModel.Instance
	if args.Get(0) != nil {
		instances = args.Get(0).([]*providerModel.Instance)
	}
	return instances, args.Error(1)
}

type instanceManagerSuite struct {
	suite.Suite
	dao     *fakeDao
	ctx     context.Context
	manager Manager
}

func (im *instanceManagerSuite) SetupSuite() {
	im.dao = &fakeDao{}
	im.manager = &manager{dao: im.dao}
	im.dao.On("List").Return(lists, nil)
}

func (im *instanceManagerSuite) TestSave() {
	im.dao.On("Create").Return(1, nil)
	id, err := im.manager.Save(im.ctx, nil)
	im.Require().Nil(err)
	im.Require().Equal(int64(1), id)
}

func (im *instanceManagerSuite) TestDelete() {
	im.dao.On("Delete").Return(nil)
	err := im.manager.Delete(im.ctx, 1)
	im.Require().Nil(err)
}

func (im *instanceManagerSuite) TestUpdate() {
	im.dao.On("Update").Return(nil)
	err := im.manager.Update(im.ctx, nil)
	im.Require().Nil(err)
}

func (im *instanceManagerSuite) TestGet() {
	ins := &providerModel.Instance{Name: "abc"}
	im.dao.On("Get").Return(ins, nil)
	res, err := im.manager.Get(im.ctx, 1)
	im.Require().Nil(err)
	im.Require().Equal(ins, res)
}

func (im *instanceManagerSuite) TestGetByName() {
	im.dao.On("GetByName").Return(lists[0], nil)
	res, err := im.manager.GetByName(im.ctx, "abc")
	im.Require().Nil(err)
	im.Require().Equal(lists[0], res)
}

func (im *instanceManagerSuite) TestCount() {
	im.dao.On("Count").Return(2, nil)
	count, err := im.manager.Count(im.ctx, nil)
	assert.Nil(im.T(), err)
	assert.Equal(im.T(), int64(2), count)
}

func (im *instanceManagerSuite) TestList() {
	lists := []*providerModel.Instance{
		{Name: "abc"},
	}
	im.dao.On("List").Return(lists, nil)
	res, err := im.manager.List(im.ctx, nil)
	assert.Nil(im.T(), err)
	assert.Len(im.T(), res, 1)
	assert.Equal(im.T(), lists, res)
}

func TestInstanceManager(t *testing.T) {
	suite.Run(t, &instanceManagerSuite{})
}

func TestEncryptAuthDataWithEmptyAuthData(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	inst := &providerModel.Instance{
		Name:     "test-instance",
		AuthData: "",
	}

	err := encryptAuthData(inst)
	require.NoError(t, err)
	assert.Equal(t, "", inst.AuthData)
}

func TestEncryptAuthDataWithData(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	authInfo := map[string]string{
		"username": "admin",
		"password": "secret123",
	}
	authData, _ := json.Marshal(authInfo)

	inst := &providerModel.Instance{
		Name:     "test-instance",
		AuthData: string(authData),
	}

	err := encryptAuthData(inst)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(inst.AuthData, utils.EncryptHeaderV1), "Encrypted data should have encryption header")
	assert.NotEqual(t, string(authData), inst.AuthData, "AuthData should be encrypted")
}

func TestDecryptAndDecodeWithEmptyAuthData(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	inst := &providerModel.Instance{
		Name:     "test-instance",
		AuthData: "",
	}

	err := decryptAndDecode(inst)
	require.NoError(t, err)
	assert.Nil(t, inst.AuthInfo)
}

func TestDecryptAndDecodeWithEncryptedData(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	authInfo := map[string]string{
		"username": "admin",
		"password": "secret123",
	}
	authData, _ := json.Marshal(authInfo)

	inst := &providerModel.Instance{
		Name:     "test-instance",
		AuthData: string(authData),
	}

	err := encryptAuthData(inst)
	require.NoError(t, err)

	err = decryptAndDecode(inst)
	require.NoError(t, err)
	assert.Equal(t, "admin", inst.AuthInfo["username"])
	assert.Equal(t, "secret123", inst.AuthInfo["password"])
}

func TestDecryptAndDecodeWithPlainJSON(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	authInfo := map[string]string{
		"username": "admin",
		"password": "secret123",
	}
	authData, _ := json.Marshal(authInfo)

	inst := &providerModel.Instance{
		Name:     "test-instance",
		AuthData: string(authData),
	}

	err := decryptAndDecode(inst)
	require.NoError(t, err)
	assert.Equal(t, "admin", inst.AuthInfo["username"])
	assert.Equal(t, "secret123", inst.AuthInfo["password"])
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(nil, kp)

	authInfo := map[string]string{
		"username":  "admin",
		"password":  "secret123",
		"api_token": "tok_abc123xyz",
	}
	authData, _ := json.Marshal(authInfo)

	inst := &providerModel.Instance{
		Name:     "test-instance",
		AuthData: string(authData),
	}

	err := encryptAuthData(inst)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(inst.AuthData, utils.EncryptHeaderV1))

	err = decryptAndDecode(inst)
	require.NoError(t, err)
	assert.Equal(t, "admin", inst.AuthInfo["username"])
	assert.Equal(t, "secret123", inst.AuthInfo["password"])
	assert.Equal(t, "tok_abc123xyz", inst.AuthInfo["api_token"])
}
