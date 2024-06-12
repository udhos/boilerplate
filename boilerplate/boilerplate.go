// Package boilerplate provides some module-wide definitions.
package boilerplate

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
)

// FuncPrintf is a helper type for logging function.
type FuncPrintf func(format string, v ...any)

// Version returns boilerplate version.
func Version() string {
	return "1.3.0"
}

// LongVersion returns boilerplate long version.
func LongVersion(me string) string {
	return fmt.Sprintf("%s runtime=%s boilerplate=%s GOOS=%s GOARCH=%s GOMAXPROCS=%d GOMEMLIMIT='%s' memory_limit=%d",
		me, runtime.Version(), Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0), os.Getenv("GOMEMLIMIT"), debug.SetMemoryLimit(-1))
}

/*
SetMemoryLimit returns the previously set memory limit.
A negative input does not adjust the limit, and allows for retrieval of the currently set memory limit.
*/
