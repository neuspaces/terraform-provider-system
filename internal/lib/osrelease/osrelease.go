package osrelease

import (
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"io"
)

// Info provides information on the operating system identification
// https://www.freedesktop.org/software/systemd/man/os-release.html
type Info struct {
	Name       string `mapstructure:"NAME,omitempty"`
	Id         string `mapstructure:"ID,omitempty"`
	PrettyName string `mapstructure:"PRETTY_NAME,omitempty"`
	Version    string `mapstructure:"VERSION,omitempty"`
	VersionId  string `mapstructure:"VERSION_ID,omitempty"`
}

func Parse(r io.Reader) (*Info, error) {
	infoMap, err := godotenv.Parse(r)
	if err != nil {
		return nil, err
	}

	var info Info
	err = mapstructure.Decode(infoMap, &info)
	if err != nil {
		return nil, err
	}

	return &info, err
}

const (
	AlpineId = "alpine"

	DebianId = "debian"

	FedoraId = "fedora"
)
