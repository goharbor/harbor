package errors

import (
	"fmt"
	"runtime"
	"strings"
)

const maxDepth = 50

type stack []uintptr

func (s *stack) frames() StackFrames {
	var stackFrames StackFrames
	frames := runtime.CallersFrames(*s)
	for {
		frame, next := frames.Next()
		// filter out runtime
		if !strings.Contains(frame.File, "runtime/") {
			stackFrames = append(stackFrames, frame)
		}
		if !next {
			break
		}
	}
	return stackFrames
}

// newStack ...
func newStack() *stack {
	var pcs [maxDepth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

// StackFrames ...
// ToDo we can define an Harbor frame to customize trace message, but it depends on requirement
type StackFrames []runtime.Frame

// Output: <File>:<Line>, <Method>
func (frames StackFrames) format() string {
	var msg string
	for _, frame := range frames {
		msg = msg + fmt.Sprintf("\n%v:%v, %v", frame.File, frame.Line, frame.Function)
	}
	return msg
}
