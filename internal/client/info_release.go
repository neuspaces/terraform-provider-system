package client

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"regexp"
	"strings"
)

type ReleaseInfo struct {
	Name    string
	Vendor  string
	Version string
	Release string
}

var (
	releasePrettyNameRegexp = regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	releaseIDRegexp         = regexp.MustCompile(`^ID=(.*)$`)
	releaseVersionIDRegexp  = regexp.MustCompile(`^VERSION_ID=(.*)$`)
	releaseUbuntuRegexp     = regexp.MustCompile(`[\( ]([\d\.]+)`)
	releaseCentOSRegexp     = regexp.MustCompile(`^CentOS( Linux)? release ([\d\.]+) `)
	releaseRedHatRegexp     = regexp.MustCompile(`[\( ]([\d\.]+)`)
)

func (c *infoClient) GetRelease(ctx context.Context) (*ReleaseInfo, error) {
	var err error

	// Retrieve contents of file `/etc/os-release`
	catCmd := &CatCommand{Path: "/etc/os-release"}
	resOsRelease, err := ExecuteCommand(ctx, c.s, catCmd)
	if err != nil {
		return nil, errors.Join(ErrInfo, err)
	}

	if resOsRelease.ExitCode != 0 || len(resOsRelease.Stdout) == 0 {
		return nil, ErrInfoUnexpected
	}

	osi := &ReleaseInfo{}

	s := bufio.NewScanner(bytes.NewReader(resOsRelease.Stdout))
	for s.Scan() {
		if m := releasePrettyNameRegexp.FindStringSubmatch(s.Text()); m != nil {
			osi.Name = strings.Trim(m[1], `"`)
		} else if m := releaseIDRegexp.FindStringSubmatch(s.Text()); m != nil {
			osi.Vendor = strings.ToLower(strings.Trim(m[1], `"`))
		} else if m := releaseVersionIDRegexp.FindStringSubmatch(s.Text()); m != nil {
			osi.Version = strings.Trim(m[1], `"`)
		}
	}

	switch osi.Vendor {
	case "debian":
		rel, err := cat(ctx, c.s, "/etc/debian_version")
		if err == nil {
			osi.Release = string(bytes.TrimSpace(rel))
		}
	case "ubuntu":
		if m := releaseUbuntuRegexp.FindStringSubmatch(osi.Name); m != nil {
			osi.Release = m[1]
		}
	case "centos":
		rel, err := cat(ctx, c.s, "/etc/centos-release")
		if err == nil {
			if m := releaseCentOSRegexp.FindStringSubmatch(string(bytes.TrimSpace(rel))); m != nil {
				osi.Release = m[2]
			}
		}
	case "rhel":
		rel, err := cat(ctx, c.s, "/etc/redhat-release")
		if err == nil {
			if m := releaseRedHatRegexp.FindStringSubmatch(string(bytes.TrimSpace(rel))); m != nil {
				osi.Release = m[1]
			}
		}
		if osi.Release == "" {
			if m := releaseRedHatRegexp.FindStringSubmatch(osi.Name); m != nil {
				osi.Release = m[1]
			}
		}
	}

	return osi, nil
}
