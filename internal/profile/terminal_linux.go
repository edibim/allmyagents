//go:build linux

package profile

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

func enableRawMode(in io.Reader, out io.Writer) (func(), bool) {
	inputFile, inputOK := in.(*os.File)
	outputFile, outputOK := out.(*os.File)
	if !inputOK || !outputOK || !isTerminal(inputFile) || !isTerminal(outputFile) {
		return func() {}, false
	}

	fd := int(inputFile.Fd())
	var original syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&original))); errno != 0 {
		return func() {}, false
	}

	raw := original
	raw.Iflag &^= syscall.ICRNL | syscall.IXON
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&raw))); errno != 0 {
		return func() {}, false
	}

	restore := func() {
		_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&original)))
	}
	return restore, true
}

func isTerminal(file *os.File) bool {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(file.Fd()), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&termios)))
	return errno == 0
}
