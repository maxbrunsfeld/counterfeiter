package terminal

import (
	"bufio"
	"os"
	"syscall"
	"unsafe"
)

type UI interface {
	TerminalIsTTY() bool
	ReadLineFromStdin() string

	Write(string)
	WriteLine(string)
}

func NewUI() UI {
	return &ui{}
}

type ui struct{}

// borrowed from github.com/golang/crypto
func isTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func (ui *ui) TerminalIsTTY() bool {
	isTTY := isTerminal(int(os.Stdin.Fd()))
	hasOverride := os.Getenv("COUNTERFEITER_INTERACTIVE") == "1"
	return isTTY || hasOverride
}

func (ui *ui) ReadLineFromStdin() string {
	bio := bufio.NewReader(os.Stdin)
	bytes, hasMoreInLine, _ := bio.ReadLine()
	line := string(bytes)

	var continuation []byte
	for hasMoreInLine {
		continuation, hasMoreInLine, _ = bio.ReadLine()
		line = line + string(continuation)
	}

	return line
}

func (ui *ui) WriteLine(line string) {
	println(line)
}

func (ui *ui) Write(output string) {
	print(output)
}
