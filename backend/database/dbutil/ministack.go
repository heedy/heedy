package dbutil

import (
	"fmt"
	"runtime"
	"strings"
)

// MiniStack generates a tiny stack trace
func MiniStack(skip int) string {
	stackSize := 3
	stack := make([]uintptr, stackSize)
	stackSize = runtime.Callers(skip+2, stack[:])

	stackString := make([]string, stackSize)
	for i := 0; i < stackSize; i++ {
		f := runtime.FuncForPC(stack[i])
		n := f.Name()
		if j := strings.LastIndex(n, "/"); j > 0 {
			n = n[j+1:]
		}
		_, line := f.FileLine(stack[i] - 1)
		stackString[i] = fmt.Sprintf("%s:%d", n, line)
	}
	return strings.Join(stackString, ";")
}
