package health_checker

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	redis "github.com/garyburd/redigo/redis"
	"github.com/streadway/amqp"
)

const (
	STATUS_OK              = iota // service of
	STATUS_CONNECTION_FAIL        // service can not be reached
	STATUS_RESULT_FAIL            // service connection ok, but ping failed
)

var (
	NotReachable = errors.New("service not reachable")
)

type CheckPoint struct {
	// name of a check point
	Name string
	// check point entrypoing, eg: localhost:6379 for redis
	DsnOrUrl string
	// underlaying connection, store db/redis/mq connection between conntion and ping  check
	// save connect time for a ping operation
	Connection interface{}

	// test if a service can be reached or not
	ServiceReachable bool

	// speed for a simple operation for specfic service
	PingSpeed time.Duration
	// if ping operation is reachable
	PingReachable bool

	// connection operation
	ConnectionChecker func() bool
	// speed testing
	SpeedChecker func() (time.Duration, error)
}

// noop connection operation, placeholder
func NoopConnectionChecker() bool {
	return true
}

// noop sppeed checker operation, placeholder
func NoopSpeedChecker() (time.Duration, error) {
	return time.Duration(0), nil
}

type HealthChecker struct {
	ServiceName string
	Status      int32
	Time        time.Duration

	CheckPoints map[string]*CheckPoint
}

func NewHealthChecker(serviceName string) *HealthChecker {
	checker := &HealthChecker{
		CheckPoints: make(map[string]*CheckPoint, 0),
		Status:      STATUS_OK,
		Time:        time.Duration(0),
		ServiceName: serviceName,
	}

	return checker
}

// add a checkpoint service to checking
// eg. Add redis checkpoint
func (checker *HealthChecker) AddCheckPoint(driver, dsnOrUrl string,
	connectionChecker func() bool,
	speedChecker func() (time.Duration, error)) {
	checkPoint := &CheckPoint{
		Name:              driver,
		DsnOrUrl:          dsnOrUrl,
		ConnectionChecker: connectionChecker,
		SpeedChecker:      speedChecker,
	}

	switch strings.ToLower(driver) {
	case "mysql":
		checkPoint.ConnectionChecker = checkPoint.MySQLConnectionChecker
		checkPoint.SpeedChecker = checkPoint.MySQLSpeedChecker
	case "redis":
		checkPoint.ConnectionChecker = checkPoint.RedisConnectionChecker
		checkPoint.SpeedChecker = checkPoint.RedisSpeedChecker
	case "mq":
		checkPoint.ConnectionChecker = checkPoint.MqConnectionChecker
		checkPoint.SpeedChecker = checkPoint.MqSpeedChecker
	}

	checker.CheckPoints[driver] = checkPoint
}

// Do checking and log result
func (checker *HealthChecker) Check() map[string]map[string]int32 {
	for name, cp := range checker.CheckPoints {
		log.Println(name)
		cp.ServiceReachable = cp.ConnectionChecker()

		speed, err := cp.SpeedChecker()
		if err != nil {
			cp.PingReachable = false
			cp.PingSpeed = time.Duration(0)
			checker.Status = STATUS_RESULT_FAIL
		} else {
			cp.PingReachable = true
			cp.PingSpeed = speed
			checker.Time += speed
		}

	}
	return checker.Response()
}

func (checker *HealthChecker) Response() map[string]map[string]int32 {
	result := make(map[string]map[string]int32, 0)
	criteria := make(map[string]int32, 0)
	criteria["Status"] = checker.Status
	criteria["Time"] = int32(checker.Time.Nanoseconds() / 1000 / 1000)
	result[checker.ServiceName] = criteria

	for name, cp := range checker.CheckPoints {
		criteria := make(map[string]int32, 0)
		if !cp.ServiceReachable {
			criteria["Status"] = STATUS_CONNECTION_FAIL
		} else if !cp.PingReachable {
			criteria["Status"] = STATUS_RESULT_FAIL
		} else {
			criteria["Status"] = STATUS_OK
		}

		criteria["Time"] = int32(cp.PingSpeed.Nanoseconds() / 1000 / 1000)
		result[name] = criteria
	}

	return result
}

func (checkPoint *CheckPoint) RedisConnectionChecker() bool {
	c, err := redis.Dial("tcp", checkPoint.DsnOrUrl)
	if err != nil {
		return false
	}

	checkPoint.Connection = c
	return true
}

func (checkPoint *CheckPoint) RedisSpeedChecker() (time.Duration, error) {
	beginTime := time.Now()
	redisConn, ok := checkPoint.Connection.(redis.Conn)
	if !ok {
		return time.Duration(-1), NotReachable
	}

	// close connection
	defer redisConn.Close()

	exists, err := redis.Bool(redisConn.Do("EXISTS", "foo"))
	log.Println(exists)
	if err != nil {
		log.Println(err)
		return time.Duration(-1), NotReachable
	}

	return time.Now().Sub(beginTime), nil
}

// https://github.com/go-sql-driver/mysql/wiki/Examples
// sql.Open doesn't make any connection
func (checkPoint *CheckPoint) MySQLConnectionChecker() bool {
	db, err := sql.Open("mysql", checkPoint.DsnOrUrl)
	if err != nil {
		log.Println(err)
		return false
	}

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return false
	}

	checkPoint.Connection = db
	return true
}

func (checkPoint *CheckPoint) MySQLSpeedChecker() (time.Duration, error) {
	beginTime := time.Now()
	db, ok := checkPoint.Connection.(*sql.DB)
	if !ok {
		return time.Duration(-1), NotReachable
	}

	// close connection
	defer db.Close()

	_, err := db.Query("SELECT now();")
	if err != nil {
		log.Println(err)
		return time.Duration(-1), NotReachable
	}

	return time.Now().Sub(beginTime), nil
}

func (checkPoint *CheckPoint) MqConnectionChecker() bool {
	conn, err := amqp.Dial(checkPoint.DsnOrUrl)
	if err != nil {
		fmt.Println(err)
		return false
	}

	checkPoint.Connection = conn
	return true
}

func (checkPoint *CheckPoint) MqSpeedChecker() (time.Duration, error) {
	beginTime := time.Now()
	conn, ok := checkPoint.Connection.(*amqp.Connection)
	if !ok {
		return time.Duration(-1), nil
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return time.Duration(-1), nil
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Println(err)
		return time.Duration(-1), nil
	}

	return time.Now().Sub(beginTime), nil
}
