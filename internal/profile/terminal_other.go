//go:build !linux

package profile

import (
	"io"
	"os"
)

func enableRawMode(_ io.Reader, _ io.Writer) (func(), bool) {
	return func() {}, false
}

func isTerminal(_ *os.File) bool {
	return false
}
