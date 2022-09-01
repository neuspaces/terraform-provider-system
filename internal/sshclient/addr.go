package sshclient

import (
	"net"
	"strconv"
)

const (
	Tcp = "tcp"
)

type hostPortAddr struct {
	net  string
	host string
	port uint16
}

func (a *hostPortAddr) Network() string {
	return a.net
}

func (a *hostPortAddr) String() string {
	return net.JoinHostPort(a.host, strconv.Itoa(int(a.port)))
}

var _ net.Addr = &hostPortAddr{}

func NewHostPortAddr(net string, host string, port uint16) net.Addr {
	return &hostPortAddr{
		net:  net,
		host: host,
		port: port,
	}
}
