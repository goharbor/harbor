package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/garyburd/redigo/redis"
	storage "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/utils"
)

const (
	namespace   = "dist_history"
	envRedisURL = "_REDIS_URL"
)

// RedisStorage implements storage based on redis backend.
type RedisStorage struct {
	redisBase *storage.RedisBase
}

// RedisStorageFactory is constructor of RedisStorage
func RedisStorageFactory() (Storage, error) {
	rawAddr := os.Getenv(envRedisURL)

	addr, ok := utils.RedisAddr(rawAddr)
	if !ok {
		return nil, fmt.Errorf("malformed redis address: %s", rawAddr)
	}
	pool := utils.RedisPool(addr)

	return &RedisStorage{
		redisBase: storage.NewRedisBase(pool, namespace),
	}, nil
}

// NewRedisStorage is constructor of RedisStorage
func NewRedisStorage(pool *redis.Pool, namespace string) *RedisStorage {
	if pool == nil || len(namespace) == 0 {
		return nil
	}

	return &RedisStorage{
		redisBase: storage.NewRedisBase(pool, namespace),
	}
}

// AppendHistory implements @Storage.AppendHistory
func (rs *RedisStorage) AppendHistory(record *models.HistoryRecord) error {
	if err := validHistoryRecord(record); err != nil {
		return err
	}

	return rs.redisBase.Save(record.TaskID, record)
}

// UpdateStatus implements @Storage.UpdateStatus
func (rs *RedisStorage) UpdateStatus(taskID string, status models.TrackStatus, startTime, endTime string) error {
	if len(taskID) == 0 {
		return errors.New("empty task ID of history record")
	}

	if !status.Valid() {
		return fmt.Errorf("invalid status %s", status)
	}

	raw, err := rs.redisBase.Get(taskID)
	if err != nil {
		return err
	}

	hr := &models.HistoryRecord{}
	if err := json.Unmarshal([]byte(raw), hr); err != nil {
		return err
	}

	hr.Status = status.String()
	if len(startTime) > 0 {
		hr.StartTime = startTime
	}
	if len(endTime) > 0 {
		hr.FinishTime = endTime
	}

	return rs.redisBase.Update(taskID, hr)
}

// LoadHistories implements @Storage.LoadHistories
func (rs *RedisStorage) LoadHistories(params *models.QueryParam) ([]*models.HistoryRecord, error) {
	rawData, err := rs.redisBase.List(params)
	if err != nil {
		return nil, err
	}

	results := []*models.HistoryRecord{}
	for _, jsonText := range rawData {
		r := &models.HistoryRecord{}
		if err := json.Unmarshal([]byte(jsonText), r); err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	return results, nil
}

func validHistoryRecord(record *models.HistoryRecord) error {
	if record == nil {
		return errors.New("nil history record")
	}

	var errs []string
	val := reflect.ValueOf(record).Elem()
	for i := 0; i < val.NumField(); i++ {
		v := val.Field(i)
		t := val.Type().Field(i)
		switch t.Type.Kind() {
		case reflect.String:
			if len(v.Interface().(string)) == 0 {
				errs = append(errs, t.Name)
			}
		case reflect.Int64:
			if v.Int() == 0 {
				errs = append(errs, t.Name)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("missing [%s]", strings.Join(errs, ","))
	}

	return nil
}
