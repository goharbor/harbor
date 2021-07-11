package oidc

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/oidc/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixEmptySubIss(t *testing.T) {
	// --- no empty subiss
	ctx := orm.Context()
	res, err := FixEmptySubIss(ctx)
	assert.False(t, res)
	assert.Nil(t, err)
	// --- create a record with empty subiss
	dao := dao.NewMetaDao()
	tok := `<enc-v1>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`
	rec := &models.OIDCUser{
		UserID: 1,
		Secret: "<enc-v1>C328t5lfBJBgBOysSUVWO4GiAM9yOFEkoX5P15Fqz1pQzKdZVGPzYN9qWcHfvt4q",
		Token:  tok,
		SubIss: "",
	}
	_, err2 := dao.Create(ctx, rec)
	require.Nil(t, err2)
	updated, err := FixEmptySubIss(ctx)
	assert.True(t, updated)
	assert.Nil(t, err)
	mgr := NewMetaMgr()
	data, err := mgr.GetByUserID(ctx, 1)
	require.Nil(t, err)
	assert.Equal(t, "CgV0ZXN0MxIFbG9jYWwhttps://10.202.250.18:5554/dex", data.SubIss)
	err3 := dao.DeleteByUserID(ctx, 1)
	require.Nil(t, err3)
}
