package cmd

import "bytes"

type BufferedCommand struct {
	Command

	StdoutBuf bytes.Buffer
	StderrBuf bytes.Buffer
}

func NewBufferedCommand(cmd string, opts ...CommandOption) *BufferedCommand {
	bc := &BufferedCommand{}
	bc.Command = NewCommand(cmd, append(opts, Stdout(&bc.StdoutBuf), Stderr(&bc.StderrBuf))...)
	return bc
}
