package nut

import (
	"net"
	"strconv"
	"time"
	"unsafe"

	nut "github.com/robbiet480/go.nut"
)

type cli struct {
	Version         string
	ProtocolVersion string
	Hostname        net.Addr
	conn            *net.TCPConn
}

func Connect(hostname string, port uint16, timeout time.Duration) (nut.Client, error) {
	if port == 0 {
		port = 3493
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(hostname, strconv.Itoa(int(port))), timeout)
	if err != nil {
		return nut.Client{}, err
	}
	client := &cli{
		Hostname: conn.RemoteAddr(),
		conn:     conn.(*net.TCPConn),
	}
	cli := (*nut.Client)(unsafe.Pointer(client))

	cli.GetVersion()
	cli.GetNetworkProtocolVersion()
	return *cli, nil
}
