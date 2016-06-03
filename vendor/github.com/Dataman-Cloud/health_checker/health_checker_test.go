package health_checker

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestNewHealthChecker(t *testing.T) {
	checker := NewHealthChecker("foobar")
	checker.AddCheckPoint("redis", "localhost:6379", nil, nil)
	checker.AddCheckPoint("mysql", "root:@/mysql", nil, nil)
	checker.AddCheckPoint("mq", "amqp://guest:guest@localhost:5672/", nil, nil)
	speedChecker := func() (time.Duration, error) {
		beginTime := time.Now()
		resp, err := http.Get("http://www.douban.com")
		if err != nil {
			log.Println(err)
			return time.Duration(-1), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Println(resp.StatusCode)
			return time.Duration(-1), nil
		}

		return time.Now().Sub(beginTime), nil
	}
	checker.AddCheckPoint("http", "www.douban.com", NoopConnectionChecker, speedChecker)

	checker.Check()
	fmt.Println(checker.Check())
}
