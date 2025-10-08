// Package diffy provides logging implementations
package diffy

import (
	"fmt"
)

type SimpleLogger struct{}

func (l *SimpleLogger) Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
