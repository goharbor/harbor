package job

var jobQueue = make(chan int64)

// Schedule put a job id into job queue.
func Schedule(jobID int64) {
	jobQueue <- jobID
}
