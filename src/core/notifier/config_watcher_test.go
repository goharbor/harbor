package notifier

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
)

var jsonText = `
{
"scan_all_policy": {
   "type": "daily",
      "parameter": {
        "daily_time": <PLACE_HOLDER>
      }
  }
}
`

func TestWatchConfiguration(t *testing.T) {
	now := time.Now().UTC()
	offset := (now.Hour()+1)*3600 + now.Minute()*60
	jsonT := strings.Replace(jsonText, "<PLACE_HOLDER>", strconv.Itoa(offset), -1)
	v := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonT), &v); err != nil {
		t.Fatal(err)
	}

	if err := WatchConfigChanges(v); err != nil {
		if !strings.Contains(err.Error(), "No handlers registered") {
			t.Fatal(err)
		}
	}
}

var jsonText2 = `
{
"scan_all_policy": {
   "type": "none"
  }
}
`

func TestWatchConfiguration2(t *testing.T) {
	v := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonText2), &v); err != nil {
		t.Fatal(err)
	}

	if err := WatchConfigChanges(v); err != nil {
		if !strings.Contains(err.Error(), "No handlers registered") {
			t.Fatal(err)
		}
	}
}
