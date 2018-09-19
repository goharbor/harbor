package chartserver

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Test the utility function parseRedisConfig
func TestParseRedisConfig(t *testing.T) {
	// Case 1: empty addr
	redisAddr := ""
	if _, err := parseRedisConfig(redisAddr); err == nil {
		t.Fatal("expect non nil error but got nil one if addr is empty")
	}

	// Case 2: short pattern, addr:port
	redisAddr = "redis:6379"
	if parsedConnStr, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatalf("expect nil error but got non nil one if addr is short pattern: %s\n", parsedConnStr)
	}

	// Case 3: long pattern but miss some parts
	redisAddr = "redis:6379,100"
	if parsedConnStr, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatalf("expect nil error but got non nil one if addr is long pattern with some parts missing: %s\n", parsedConnStr)
	} else {
		if strings.Index(parsedConnStr, `"dbNum":"0"`) == -1 {
			t.Fatalf("expect 'dbNum:0' in the parsed conn str but got nothing: %s\n", parsedConnStr)
		}
	}

	// Case 4: long pattern
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

func TestGetCacheConfig(t *testing.T) {
	// case 1: no cache set
	cacheConf, err := getCacheConfig()
	if err != nil || cacheConf != nil {
		t.Fatal("expect nil cache config and nil error but got non-nil one when parsing empty cache settings")
	}

	// case 2: unknown cache type
	os.Setenv(cacheDriverENVKey, "unknown")
	_, err = getCacheConfig()
	if err == nil {
		t.Fatal("expect non-nil error but got nil one when parsing unknown cache type")
	}

	// case 3: in memory cache type
	os.Setenv(cacheDriverENVKey, cacheDriverMem)
	memCacheConf, err := getCacheConfig()
	if err != nil || memCacheConf == nil || memCacheConf.DriverType != cacheDriverMem {
		t.Fatal("expect in memory cache driver but got invalid one")
	}

	// case 4: wrong redis cache conf
	os.Setenv(cacheDriverENVKey, cacheDriverRedis)
	os.Setenv(redisENVKey, "")
	_, err = getCacheConfig()
	if err == nil {
		t.Fatal("expect non-nil error but got nil one when parsing a invalid redis cache conf")
	}

	// case 5: redis cache conf
	os.Setenv(redisENVKey, "redis:6379,100,Passw0rd,1")
	redisConf, err := getCacheConfig()
	if err != nil {
		t.Fatalf("expect nil error but got non-nil one when parsing valid redis conf")
	}

	if redisConf == nil || redisConf.DriverType != cacheDriverRedis {
		t.Fatal("expect redis cache driver but got invalid one")
	}

	conf := make(map[string]string)
	if err = json.Unmarshal([]byte(redisConf.Config), &conf); err != nil {
		t.Fatal(err)
	}

	if v, ok := conf["conn"]; !ok {
		t.Fatal("expect 'conn' filed in the parsed conf but got nothing")
	} else {
		if v != "redis:6379" {
			t.Fatalf("expect %s but got %s", "redis:6379", v)
		}
	}

	// clear
	os.Unsetenv(cacheDriverENVKey)
	os.Unsetenv(redisENVKey)
}
