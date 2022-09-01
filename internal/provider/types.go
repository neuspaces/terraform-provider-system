package provider

import (
	"fmt"
	"io/fs"
	"strconv"
)

type Mode uint32

func (m Mode) String() string {
	return fmt.Sprintf("%o", m)
}

func parseMode(mode string) (fs.FileMode, error) {
	m, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0, err
	}
	return fs.FileMode(m), err
}

func mustParseMode(mode string) fs.FileMode {
	m, err := parseMode(mode)
	if err != nil {
		return 0
	}
	return m
}
