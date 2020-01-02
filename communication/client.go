package communication

import (
	"bufio"
	"io"
	"syscall"
)

type client struct {
	// Socket descriptor
	socket int
	// Ipv4
	ip string
	// TCP port
	port int
	// Raw Socket address
	address syscall.Sockaddr
	// Clients buffer
	buffer *bufio.ReadWriter
	reader *io.PipeReader
	writer *io.PipeWriter
}
