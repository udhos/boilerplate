// Package boilerplate provides some module-wide definitions.
package boilerplate

import (
	"fmt"
	"runtime"
)

// FuncPrintf is a helper type for logging function.
type FuncPrintf func(format string, v ...any)

// Version returns boilerplate version.
func Version() string {
	return "0.10.0"
}

// LongVersion returns boilerplate long version.
func LongVersion(me string) string {
	return fmt.Sprintf("%s runtime=%s boilerplate=%s GOOS=%s GOARCH=%s GOMAXPROCS=%d",
		me, runtime.Version(), Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))
}
