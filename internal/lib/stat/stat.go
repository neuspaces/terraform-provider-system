package stat

import (
	"io/fs"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var regexpSymlinkName = regexp.MustCompile(`^'([^']*)' -> '([^']*)'$`)

type Stat struct {
	Platform     string
	Mode         FileMode
	Name         string
	Target       string
	User         string
	Uid          int
	Group        string
	Gid          int
	Size         int64
	AccessTime   time.Time
	ModifiedTime time.Time
	ChangeTime   time.Time
}

// ToFsFileInfo returns the Stat as a fs.FileInfo
func (s *Stat) ToFsFileInfo() fs.FileInfo {
	return &statFileInfo{
		stat: *s,
	}
}

func parseUnixEpochUTC(epoch string) (time.Time, error) {
	epochInt, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(epochInt, 0), nil
}

// unquoteSingle removes single quotes from the start and end of the provided string.
// unquoteSingle returns the provided string unchanged when there are no single quotes at the start and end.
func unquoteSingle(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		return s[1 : len(s)-1]
	}
	return s
}

// statFileInfo wraps a Stat and implements the fs.FileInfo interface.
type statFileInfo struct {
	stat Stat
}

var _ fs.FileInfo = &statFileInfo{}

func (s *statFileInfo) Name() string {
	return s.stat.Name
}

func (s *statFileInfo) Size() int64 {
	return s.stat.Size
}

func (s *statFileInfo) Mode() fs.FileMode {
	return s.stat.Mode.ToFsFileMode()
}

func (s *statFileInfo) ModTime() time.Time {
	return s.stat.ChangeTime
}

func (s *statFileInfo) IsDir() bool {
	return s.stat.Mode.IsDir()
}

func (s *statFileInfo) Sys() interface{} {
	return nil
}
