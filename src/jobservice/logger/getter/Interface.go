package getter

// Interface defines operations of a log data getter
type Interface interface {
	// Retrieve the log data of the specified log entry
	//
	// logID string : the id of the log entry. e.g: file name a.log for file log
	//
	// If succeed, log data bytes will be returned
	// otherwise, a non nil error is returned
	Retrieve(logID string) ([]byte, error)
}
