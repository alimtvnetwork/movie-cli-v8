// stack.go — stack trace capture helper for AppError.
package appfault

import (
	"path/filepath"
	"runtime"
	"strings"
)

const defaultStackDepth = 15

func captureStackTrace(skip int) StackTrace {
	var frames StackTrace

	for i := skip; i < skip+defaultStackDepth; i++ {
		pc, file, line, isOK := runtime.Caller(i)

		if !isOK {
			break
		}

		fnName := "unknown"
		fnObj := runtime.FuncForPC(pc)

		if fnObj != nil {
			parts := strings.Split(fnObj.Name(), "/")
			fnName = parts[len(parts)-1]
		}

		frames = append(frames, StackFrame{
			Function: fnName,
			File:     filepath.Base(file),
			Line:     line,
		})
	}

	return frames
}
