package job

var JobQueue chan int64 = make(chan int64)

func Schedule(jobID int64) {
	JobQueue <- jobID
}
