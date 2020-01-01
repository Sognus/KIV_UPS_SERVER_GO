package communication

import "syscall"

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
	buffer chan byte
}
