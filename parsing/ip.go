package parsing

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"syscall"
)

// Parses port from string to unsigned char
func ParsePort(port string) (int, error) {
	val, err := strconv.ParseUint(port, 10, 16)

	// Check for parse error
	if err != nil {
		return (int)(val), err
	}

	// Return parsed val
	return (int)(val), nil
}

func AddressFromBytes(bytes []byte, port string) (syscall.Sockaddr, error) {
	// Declaration
	var err error = nil
	var address syscall.Sockaddr = nil

	// Check if bytes len is 4
	if len(bytes) != 4 {
		text := fmt.Sprintf("communications: Expected 4 bytes - %d given", len(bytes))
		return nil, errors.New(text)
	}

	parsedPort, parseErr := ParsePort(port)

	if parseErr != nil {
		return nil, parseErr
	}

	// Convert slice to fixed size array
	var addrBytes [4]byte
	copy(addrBytes[:], bytes[:4])


	// Create IPv4 address from bytes and parsedPort
	address = &syscall.SockaddrInet4{
		Port: parsedPort,
		Addr: addrBytes,
	}

	// Return result
	return address, err
}

func AddressFromString(ip string, port string) (syscall.Sockaddr, error) {
	// Create storage for IP string parsing
	bytes := make([]byte, 4, 4)

	re, err := regexp.Compile(`(\d+)\.(\d+)\.(\d+)\.(\d+)`)

	// Regex error handling
	if err != nil {
		return nil, err
	}

	// Find values from string
	parsedIP := re.FindStringSubmatch(ip)

	// Check if we have enough of them
	if len(parsedIP) != 5 {
		return nil, errors.New("given string is not an IPv4 adress")
	}

	// Parse found element to 8bit unsigned int
	for index, element := range parsedIP[1:5] {
		data, err := strconv.ParseUint(element, 10, 8)

		if err != nil {
			return nil, err
		}

		bytes[index] = byte(data)
	}

	// Create SockAddr from parsed bytes
	return AddressFromBytes(bytes, port)
}
