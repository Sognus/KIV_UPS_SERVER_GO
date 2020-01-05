package communication

import (
	"io"
	"syscall"
)

type Client struct {
	UID int
	// Socket descriptor
	Socket int
	// Ipv4
	ip string
	// TCP port
	port int
	// Raw Socket address
	address syscall.Sockaddr
	// Clients Reader and Writer
	Reader *io.PipeReader
	writer *io.PipeWriter
}
