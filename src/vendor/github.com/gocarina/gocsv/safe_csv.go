package gocsv

//Wraps around SafeCSVWriter and makes it thread safe.
import (
	"encoding/csv"
	"sync"
)

type CSVWriter interface {
	Write(row []string) error
	Flush()
	Error() error
}

type SafeCSVWriter struct {
	*csv.Writer
	m sync.Mutex
}

func NewSafeCSVWriter(original *csv.Writer) *SafeCSVWriter {
	return &SafeCSVWriter{
		Writer: original,
	}
}

//Override write
func (w *SafeCSVWriter) Write(row []string) error {
	w.m.Lock()
	defer w.m.Unlock()
	return w.Writer.Write(row)
}

//Override flush
func (w *SafeCSVWriter) Flush() {
	w.m.Lock()
	w.Writer.Flush()
	w.m.Unlock()
}
