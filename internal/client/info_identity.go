package client

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

type IdentityInfo struct {
	Name  string
	Uid   int
	Group string
	Gid   int
}

var (
	identityIdRegexp = regexp.MustCompile(`^(?:uid=(?P<uid>\d+)(?:\((?P<user>\w+)\))?)?\s*(?:gid=(?P<gid>\d+)(?:\((?P<group>\w+)\))?)?\s*(?:groups=(?P<groups>(?:\d+\(.*\))*))?$`)
)

func (c *infoClient) GetIdentity(ctx context.Context) (*IdentityInfo, error) {
	var err error

	resId, err := ExecuteCommand(ctx, c.s, SimpleCommand(`id`))
	if err != nil {
		return nil, ErrInfo.Raise(err)
	}

	if resId.ExitCode != 0 || len(resId.Stdout) == 0 {
		return nil, ErrInfoUnexpected
	}

	idMatch := identityIdRegexp.FindStringSubmatch(strings.TrimSpace(resId.StdoutString()))
	if idMatch == nil || len(idMatch) != 6 {
		return nil, ErrInfoUnexpected
	}

	uid, err := strconv.Atoi(idMatch[1])
	if err != nil {
		return nil, ErrServiceUnexpected
	}

	gid, err := strconv.Atoi(idMatch[3])
	if err != nil {
		return nil, ErrServiceUnexpected
	}

	ui := &IdentityInfo{
		Name:  idMatch[2],
		Uid:   uid,
		Group: idMatch[4],
		Gid:   gid,
	}

	return ui, err
}
