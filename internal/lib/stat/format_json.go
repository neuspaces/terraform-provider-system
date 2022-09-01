package stat

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	FormatJsonGnu = `{"plat":"gnu","mode":"%f","name":"%N","user":"%U","uid":"%u","group":"%G","gid":"%g","size":"%s","atime":"%X","mtime":"%Y","ctime":"%Z"}`
)

type jsonFormat struct {
	Platform     string `json:"plat"`
	Mode         string `json:"mode"`
	Name         string `json:"name"`
	User         string `json:"user"`
	Uid          string `json:"uid"`
	Group        string `json:"group"`
	Gid          string `json:"gid"`
	Size         string `json:"size"`
	AccessTime   string `json:"atime"`
	ModifiedTime string `json:"mtime"`
	ChangedTime  string `json:"ctime"`
}

func unmarshalJsonFormat(data []byte) (*jsonFormat, error) {
	var sj jsonFormat
	err := json.Unmarshal(data, &sj)
	if err != nil {
		return nil, err
	}
	return &sj, nil
}

// ParseJsonFormat parsed the provided output of `stat` which is formatted using FormatJsonGnu.
func ParseJsonFormat(data []byte) (*Stat, error) {
	us, err := unmarshalJsonFormat(data)
	if err != nil {
		return nil, err
	}

	// Mode (from hex) %f
	// Stat provides mode in 2 byte = 16 bits hex
	modeInt, err := strconv.ParseUint(us.Mode, 16, 16)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse mode '%s'", us.Mode))
	}

	mode := FileMode(modeInt)

	// Name and target
	var name string
	var target string

	if mode.IsSymlink() {
		// If type is link, the "name" field has the form 'name' -> 'link' (example: "'link-to-target.txt' -> '/root/target.txt'")
		nameTargetMatch := regexpSymlinkName.FindStringSubmatch(us.Name)
		if nameTargetMatch == nil || len(nameTargetMatch) != 3 {
			return nil, newParseError(fmt.Sprintf("unexpected name '%s' for symlink", us.Name))
		}

		name = nameTargetMatch[1]
		target = nameTargetMatch[2]
	} else {
		// stat MAY return the name in single quotes
		name = unquoteSingle(us.Name)
	}

	// Size %s
	size, err := strconv.ParseInt(us.Size, 10, 64)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse size %s", us.Size))
	}

	// Uid %u
	uid, err := strconv.Atoi(us.Uid)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse uid %s", us.Uid))
	}

	// Gid %g
	gid, err := strconv.Atoi(us.Gid)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse gid %s", us.Gid))
	}

	// Access time %X
	accessTime, err := parseUnixEpochUTC(us.AccessTime)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse access time %s", us.AccessTime))
	}

	// Modified time %Y
	modifiedTime, err := parseUnixEpochUTC(us.ModifiedTime)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse modified time %s", us.ModifiedTime))
	}

	// Change time %Z
	changeTime, err := parseUnixEpochUTC(us.ChangedTime)
	if err != nil {
		return nil, newParseError(fmt.Sprintf("failed to parse change time %s", us.ChangedTime))
	}

	s := &Stat{
		Platform:     us.Platform,
		Mode:         mode,
		Name:         name,
		Target:       target,
		User:         us.User,
		Uid:          uid,
		Group:        us.Group,
		Gid:          gid,
		Size:         size,
		AccessTime:   accessTime,
		ModifiedTime: modifiedTime,
		ChangeTime:   changeTime,
	}

	return s, nil
}
