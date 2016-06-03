# Health Check Services

  ```
	  checker := NewHealthChecker()
  ```
## Redis

  ```
	checker.AddCheckPoint("redis", "localhost:6379", nil, nil)
  ```

## MySQL
  ```
	checker.AddCheckPoint("mysql", "root:@/mysql", nil, nil)
  ```

## Mq
  ```
	checker.AddCheckPoint("mq", "amqp://guest:guest@localhost:5672/", nil, nil)
  ```

## HTTP
  ```
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
  ```


## Fire Check

  ```
    checker.Check()
  ```

## Result

  ```
    {"http":{"Status":0,"Time":150404704},"mq":{"Status":0,"Time":850894},"mysql":{"Status":0,"Time":102477},"redis":{"Status":0,"Time":119484}}
  ```
  
