package instance

import (
	"fmt"
	"testing"
	"time"

	storage "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/tests"
)

var (
	pool       = tests.Pool()
	testingKey = "test_instance_storage"
)

func TestDel(t *testing.T) {
	rs := NewRedisStorage(pool, testingKey)
	if rs == nil {
		t.Fatal("expect non nil redis storage but got nil")
	}
	defer tests.Clear(pool, testingKey)

	id, err := mockOne(rs)
	if err != nil {
		t.Fatal(err)
	}

	if err := rs.Delete(id); err != nil {
		t.Error(err)
	}

	if err := rs.Delete("not-existing"); err != nil {
		if err != storage.ErrObjectNotFound {
			t.Errorf("expect ErrObjectNotFound but got %s", err)
		}
	} else {
		t.Error("expect non nil error but got nil")
	}
}

func TestUpdate(t *testing.T) {
	rs := NewRedisStorage(pool, testingKey)
	if rs == nil {
		t.Fatal("expect non nil redis storage but got nil")
	}
	defer tests.Clear(pool, testingKey)

	id, err := mockOne(rs)
	if err != nil {
		t.Fatal(err)
	}

	meta, err := rs.Get(id)
	if err != nil {
		t.Fatal(err)
	}

	meta.Endpoint = "http://127.0.0.1"
	if err := rs.Update(meta); err != nil {
		t.Fatal(err)
	}

	meta, err = rs.Get(id)
	if err != nil {
		t.Fatal(err)
	}

	if meta.Endpoint != "http://127.0.0.1" {
		t.Errorf("expect endpoint 'http://127.0.0.1' but got %s", meta.Endpoint)
	}

}

func TestList(t *testing.T) {
	rs := NewRedisStorage(pool, testingKey)
	if rs == nil {
		t.Fatal("expect non nil redis storage but got nil")
	}
	defer tests.Clear(pool, testingKey)

	_, err := mockOne(rs)
	if err != nil {
		t.Fatal(err)
	}

	_, err = mockOne(rs)
	if err != nil {
		t.Fatal(err)
	}

	l, err := rs.List(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(l) != 2 {
		t.Errorf("expect 2 instance metadata but got %d", len(l))
	}
}

func mockOne(rs *RedisStorage) (string, error) {
	fakeObj := giveMeMetadata()
	id, err := rs.Save(fakeObj)
	if err != nil {
		return "", err
	}

	if len(id) == 0 {
		return "", fmt.Errorf("expect id but got empty")
	}

	return id, nil
}

func giveMeMetadata() *models.Metadata {
	return &models.Metadata{
		Name:           "us-east",
		Description:    "for UT",
		Provider:       "Dragonfly",
		Endpoint:       "http://localhost/endpoint",
		AuthMode:       "BASIC",
		AuthData:       map[string]string{"admin": "Passw0rd"},
		Status:         "Healthy",
		Enabled:        true,
		SetupTimestamp: time.Now().Unix(),
	}
}
