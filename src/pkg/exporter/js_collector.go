package exporter

import (
	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
)

// JobServiceCollectorName ...
const JobServiceCollectorName = "JobServiceCollector"

var (
	jobServiceTaskQueueSize = typedDesc{
		desc:      newDescWithLables("", "task_queue_size", "Total number of tasks", "type"),
		valueType: prometheus.GaugeValue,
	}
	jobServiceTaskQueueLatency = typedDesc{
		desc:      newDescWithLables("", "task_queue_latency", "how long ago the next job to be processed was enqueued", "type"),
		valueType: prometheus.GaugeValue,
	}
	jobServiceConcurrency = typedDesc{
		desc:      newDescWithLables("", "task_concurrency", "Total number of concurrency on a pool", "type", "pool"),
		valueType: prometheus.GaugeValue,
	}
	jobServiceScheduledJobTotal = typedDesc{
		desc:      newDesc("", "task_scheduled_total", "total number of scheduled job"),
		valueType: prometheus.GaugeValue,
	}
)

// NewJobServiceCollector ...
func NewJobServiceCollector() *JobServiceCollector {
	return &JobServiceCollector{Namespace: namespace}
}

// JobServiceCollector ...
type JobServiceCollector struct {
	Namespace string
}

// Describe implements prometheus.Collector
func (hc *JobServiceCollector) Describe(c chan<- *prometheus.Desc) {
	for _, jd := range hc.getDescribeInfo() {
		c <- jd
	}
}

// Collect implements prometheus.Collector
func (hc *JobServiceCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range hc.getJobserviceInfo() {
		c <- m
	}
}

// GetName returns the name of the job service collector
func (hc *JobServiceCollector) GetName() string {
	return JobServiceCollectorName
}

func (hc *JobServiceCollector) getDescribeInfo() []*prometheus.Desc {
	return []*prometheus.Desc{
		jobServiceTaskQueueSize.Desc(),
		jobServiceTaskQueueLatency.Desc(),
		jobServiceConcurrency.Desc(),
		jobServiceScheduledJobTotal.Desc(),
	}
}

func (hc *JobServiceCollector) getJobserviceInfo() []prometheus.Metric {
	if CacheEnabled() {
		value, ok := CacheGet(JobServiceCollectorName)
		if ok {
			return value.([]prometheus.Metric)
		}
	}

	// Get concurrency info via raw redis client
	result := getConccurrentInfo()

	// get info via jobservice client
	cli := GetBackendWorker()
	// get queue info
	qs, err := cli.Queues()
	checkErr(err, "error when get work task queues info")
	for _, q := range qs {
		result = append(result, jobServiceTaskQueueSize.MustNewConstMetric(float64(q.Count), q.JobName))
		result = append(result, jobServiceTaskQueueLatency.MustNewConstMetric(float64(q.Latency), q.JobName))
	}

	// get scheduled job info
	_, total, err := cli.ScheduledJobs(0)
	checkErr(err, "error when get scheduled job number")
	result = append(result, jobServiceScheduledJobTotal.MustNewConstMetric(float64(total)))

	if CacheEnabled() {
		CachePut(JobServiceCollectorName, result)
	}
	return result
}

func getConccurrentInfo() []prometheus.Metric {
	rdsConn := GetRedisPool().Get()
	defer rdsConn.Close()
	result := []prometheus.Metric{}
	knownJobvalues, err := redis.Values(rdsConn.Do("SMEMBERS", redisKeyKnownJobs(jsNamespace)))
	checkErr(err, "err when get known jobs")
	for _, v := range knownJobvalues {
		job := string(v.([]byte))
		lockInfovalues, err := redis.Values(rdsConn.Do("HGETALL", redisKeyJobsLockInfo(jsNamespace, job)))
		checkErr(err, "err when get job lock info")
		for i := 0; i < len(lockInfovalues); i += 2 {
			key, _ := redis.String(lockInfovalues[i], nil)
			value, _ := redis.Float64(lockInfovalues[i+1], nil)
			result = append(result, jobServiceConcurrency.MustNewConstMetric(value, job, key))
		}
	}
	return result
}
