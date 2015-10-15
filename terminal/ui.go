package terminal

import (
	"bufio"
	"os"

	"golang.org/x/crypto/ssh/terminal"
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

func (ui *ui) TerminalIsTTY() bool {
	return terminal.IsTerminal(int(os.Stdin.Fd()))
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
