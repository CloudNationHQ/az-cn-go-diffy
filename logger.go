// Package diffy provides logging implementations
package diffy

import (
	"fmt"
)

// SimpleLogger implements Logger interface.
type SimpleLogger struct{}

func (l *SimpleLogger) Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
