package history

import (
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/tests"
)

var (
	pool       = tests.Pool()
	testingKey = "test_history_storage"
)

func TestAppendHistory(t *testing.T) {
	rStorage := NewRedisStorage(pool, testingKey)
	if rStorage == nil {
		t.Fatal("expect non nil history redis storage object but got nil")
	}

	defer tests.Clear(pool, testingKey)

	if err := rStorage.AppendHistory(giveMeEmptyHistory()); err == nil {
		t.Fatal("expect non nil error but got nil when append invalid history")
	}

	if err := rStorage.AppendHistory(giveMeHistory()); err != nil {
		t.Fatalf("expect nil error but got %s when append valid history", err)
	}
}

func TestLoadHistory(t *testing.T) {
	rStorage := NewRedisStorage(pool, testingKey)
	if rStorage == nil {
		t.Fatal("expect non nil history redis storage object but got nil")
	}

	defer tests.Clear(pool, testingKey)
	for i := 0; i < 26; i++ {
		if err := rStorage.AppendHistory(giveMeHistory()); err != nil {
			t.Fatalf("append history failed with error: %s", err)
		}
	}

	records, err := rStorage.LoadHistories(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 25 {
		t.Fatalf("expect 25 history records with nil query param but got %d", len(records))
	}

	records, err = rStorage.LoadHistories(&models.QueryParam{
		Page:     3,
		PageSize: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("expect 0 records in page 3 but got %d", len(records))
	}

	records, err = rStorage.LoadHistories(&models.QueryParam{
		Page:     2,
		PageSize: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expect 1 records in page 2 but got %d", len(records))
	}

	records, err = rStorage.LoadHistories(&models.QueryParam{
		Page:     1,
		PageSize: 25,
		Keyword:  "steven",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("expect 0 records in page 1 with keyword 'steven' but got %d", len(records))
	}
}

func TestUpdateHistory(t *testing.T) {
	rStorage := NewRedisStorage(pool, testingKey)
	if rStorage == nil {
		t.Fatal("expect non nil history redis storage object but got nil")
	}

	defer tests.Clear(pool, testingKey)

	testingObj := giveMeHistory()
	if err := rStorage.AppendHistory(testingObj); err != nil {
		t.Fatalf("expect nil error but got %s when append valid history", err)
	}

	if err := rStorage.UpdateStatus("task_ID_1", models.PreheatingStatusFail, "", ""); err != nil {
		t.Fatal(err)
	}

	items, err := rStorage.LoadHistories(nil)
	if err != nil {
		t.Fatal(err)
	}

	if items[0].Status != models.PreheatingStatusFail {
		t.Errorf("expect status '%s' but got '%s'", models.PreheatingStatusFail, items[0].Status)
	}

}

func giveMeHistory() *models.HistoryRecord {
	t := time.Now().UnixNano()
	return &models.HistoryRecord{
		TaskID:     "task_ID_1",
		Image:      fmt.Sprintf("image_%d", t),
		StartTime:  "-",
		FinishTime: "-",
		Status:     "SUCCESS",
		Provider:   "Dragonfly",
		Instance:   fmt.Sprintf("inst_id_%d", t),
	}
}

func giveMeEmptyHistory() *models.HistoryRecord {
	return &models.HistoryRecord{}
}
