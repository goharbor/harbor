package history

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models"
)

func validHistoryRecord(record *models.HistoryRecord) error {
	if record == nil {
		return errors.New("nil history record")
	}

	var errs []string
	typ := reflect.TypeOf(*record)
	val := reflect.ValueOf(record).Elem()
	for i := 0; i < val.NumField(); i++ {
		if typ.Field(i).Name == "ID" {
			continue
		}

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
