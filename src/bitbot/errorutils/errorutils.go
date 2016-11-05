package errorutils

import (
	"log"
	"runtime"
)

// LogPanic logs a formatted stack trace of the panicing goroutine. The stack trace is truncated
// at 4096 bytes (https://groups.google.com/d/topic/golang-nuts/JGraQ_Cp2Es/discussion)
func LogPanic() {
	if err := recover(); err != nil {
		const size = 4096
		buf := make([]byte, size)
		stack := buf[:runtime.Stack(buf, false)]
		log.Printf("Error: %v\n%s", err, stack)
	}
}

// PanicOnError panic if the given err is not nil.
func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
