package chartserver

import (
	"strings"
	"testing"
)

//Test the utility function parseRedisConfig
func TestParseRedisConfig(t *testing.T) {
	//Case 1: empty addr
	redisAddr := ""
	if _, err := parseRedisConfig(redisAddr); err == nil {
		t.Fatal("expect non nil error but got nil one if addr is empty")
	}

	//Case 2: short pattern, addr:port
	redisAddr = "redis:6379"
	if parsedConnStr, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatalf("expect nil error but got non nil one if addr is short pattern: %s\n", parsedConnStr)
	}

	//Case 3: long pattern but miss some parts
	redisAddr = "redis:6379,100"
	if parsedConnStr, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatalf("expect nil error but got non nil one if addr is long pattern with some parts missing: %s\n", parsedConnStr)
	} else {
		if strings.Index(parsedConnStr, `"dbNum":"0"`) == -1 {
			t.Fatalf("expect 'dbNum:0' in the parsed conn str but got nothing: %s\n", parsedConnStr)
		}
	}

	//Case 4: long pattern
	redisAddr = "redis:6379,100,Passw0rd,1"
	if parsedConnStr, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatal("expect nil error but got non nil one if addr is long pattern")
	} else {
		if strings.Index(parsedConnStr, `"dbNum":"1"`) == -1 ||
			strings.Index(parsedConnStr, `"password":"Passw0rd"`) == -1 {
			t.Fatalf("expect 'dbNum:0' and 'password:Passw0rd' in the parsed conn str but got nothing: %s", parsedConnStr)
		}
	}
}
