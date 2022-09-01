package stat

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// FormatTerseGnu
	FormatTerseGnu = `%n %s %b %f %u %g %D %i %h %t %T %X %Y %Z %o`
)

func ParseTerseFormat(data []byte) (*Stat, error) {
	d := string(data)

	// split space-separated parts
	parts := strings.Split(d, " ")
	lenParts := len(parts)

	// expect 15 parts
	// if there are less than 15 parts, the format is invalid
	if lenParts < 15 {
		return nil, newParseError("invalid format")
	}

	// if there are more than 15 parts, the file name includes spaces
	// recombine to exactly 15 parts
	if lenParts > 15 {
		newParts := []string{strings.Join(parts[0:lenParts-15+1], " ")}
		newParts = append(newParts, parts[lenParts-15+1:lenParts]...)
		parts = newParts
	}

	name := parts[0]

	// Size %s
	size, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse size %s", parts[1]))
	}

	// Mode (from hex) %f
	modeInt, err := strconv.ParseUint(parts[3], 16, 32)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse mode '%s'", parts[3]))
	}

	mode := FileMode(modeInt)

	// Access time %X
	accessTime, err := parseUnixEpochUTC(parts[11])
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse access time %s", parts[11]))
	}

	// Modified time %Y
	modifiedTime, err := parseUnixEpochUTC(parts[12])
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse modified time %s", parts[12]))
	}

	// Change time %Z
	changeTime, err := parseUnixEpochUTC(parts[13])
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse change time %s", parts[13]))
	}

	s := &Stat{
		Mode:         mode,
		Name:         name,
		Size:         size,
		AccessTime:   accessTime,
		ModifiedTime: modifiedTime,
		ChangeTime:   changeTime,
	}

	return s, nil
}
