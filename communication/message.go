package communication

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	startCharacter = '<'
	endCharacter   = '>'
	escapeCharacter = '\\'
	valueDelimiter = ':'
	pairDelimiter  = ';'
	headEnd        = '|'
)

type message struct {
	// Message ID
	id int
	// Message return ID
	rid int
	// Message type
	msg int
	// Message content
	content map[string]string

}

// Returns if byte is control byte
func isControl(character byte) bool {
	if character == startCharacter {
		return true
	}
	if character == endCharacter {
		return true
	}
	if character == escapeCharacter {
		return true
	}
	if character == valueDelimiter {
		return true
	}
	if character == pairDelimiter {
		return true
	}
	if character == headEnd {
		return true
	}
	return false
}

func Decoder(client *Client) {
	buffer := bufio.NewReader(client.Reader)

	// Declare values
	var errReadHeader error
	var errReadValue error
	var valueInt int

	//


	for {
		var msg message
		errStart := WaitForStart(buffer)


		// Ignore error
		if errStart != nil {
			continue
		}

		// Read bytes until  "id:" is read
		_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "id")

		if errReadHeader != nil {
			continue
		}

		// Read bytes for integer value
		valueInt, errReadValue = ReadPairValueInt(buffer)

		if errReadValue != nil {
			continue
		}

		// Set message ID
		msg.id = valueInt

		// Read bytes until  "rid:" is read
		_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "rid")

		if errReadHeader != nil {
			continue
		}

		// Read bytes for integer value
		valueInt, errReadValue = ReadPairValueInt(buffer)

		if errReadValue != nil {
			continue
		}

		// Set message return ID
		msg.rid = valueInt

		// Read bytes until "type:" is read
		_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "type")

		if errReadHeader != nil {
			continue
		}

		// Read bytes for integer value
		valueInt, errReadValue = ReadPairValueInt(buffer)

		if errReadValue != nil {
			continue
		}

		// Set message type
		msg.msg = valueInt

		// Expect header end byte
		errExpect := ExpectByte(buffer, headEnd)

		if errExpect != nil {
			continue
		}

		// Read message content
		msg.content = make(map[string]string)


		for {
			headerData, errReadHeader := ReadPairHeaderAny(buffer, valueDelimiter)

			if errReadHeader != nil {
				break
			}

			fmt.Printf("PairHeader: %s\n", headerData)

			valueData, errReadValue := ReadPairValueString(buffer)

			if errReadValue != nil {
				break
			}

			fmt.Printf("PairValue: %s\n", valueData)

			fmt.Printf("msg.content[%s] = %s\n", headerData, valueData)
			msg.content[string(headerData)] = valueData

			errPeek := PeekExpectByte(buffer, endCharacter)


			if errPeek != nil {
				continue
			} else {
				break
			}
		}

		fmt.Printf("#%d > Message: %v\n", client.Socket, msg)

	}
}

// Reads bytes until start character is read
func WaitForStart(reader io.Reader) error {
	for {
		startBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, startBuffer)

		if errRead != nil {
			return errRead
		}

		if startBuffer[0] == startCharacter {
			return nil
		}
	}
}

// Checks next byte without consuming
func PeekExpectByte(reader *bufio.Reader, expect byte) error {
	expectBuffer, errRead := reader.Peek(1)

	if errRead != nil {
		return errRead
	}

	if expectBuffer[0] == expect {
		return nil
	} else {
		return errors.New("wrong byte value")
	}
}


func ExpectByte(reader io.Reader, expected byte) error {
	expectBuffer := make([]byte, 1)
	_, errRead := io.ReadFull(reader, expectBuffer)

	if errRead != nil {
		return errRead
	}

	if expectBuffer[0] == expected {
		return nil
	}

	return errors.New("wrong byte read")

}

// Similar to ReadPairHeader without match requirement
func ReadPairHeaderAny(reader io.Reader, character byte) ([]byte, error) {
	var buffer []byte = make([]byte, 0, 0)
	var escape bool = false

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return nil, errRead
		}

		if escape {
			if characterBuffer[0] == escapeCharacter {
				buffer = append(buffer, characterBuffer[0])
				escape = false
			} else {
				if isControl(characterBuffer[0]) {
					buffer = append(buffer, characterBuffer[0])
					escape = false
				} else {
					return nil, errors.New("control byte was expected")
				}
			}
		} else {
			if characterBuffer[0] == escapeCharacter {
				escape = true
			} else {
				if characterBuffer[0] == character {
					return buffer, nil
				} else {
					if isControl(characterBuffer[0]) {
						return nil, errors.New("unexpected control byte")
					} else {
						buffer = append(buffer, characterBuffer[0])
					}
				}
			}
		}
	}
}

// Reads bytes and detects given header in message
func ReadPairHeader(reader io.Reader, character byte, expected string) ([]byte, error) {
	var buffer []byte = make([]byte, 0, 0)

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return nil, errRead
		}

		// When we reach delimiter character we stop
		if characterBuffer[0] == character {
			currentData := fmt.Sprintf("%s", buffer)
			if currentData != expected || len(currentData) > len(expected) {
				return nil, errors.New("unexpected string")
			} else {
				return buffer, nil
			}
		} else {
			buffer = append(buffer, characterBuffer[0])
		}
	}
}

func ReadPairValueInt(reader io.Reader) (int, error) {
	var buffer []byte = make([]byte, 0, 0)

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return 0, errRead
		}

		if characterBuffer[0] == pairDelimiter {
			// Parse data to int
			str := fmt.Sprintf("%s", buffer)
			number, errNumber := strconv.Atoi(str)

			if errNumber != nil {
				return 0, errNumber
			} else {
				return number, nil
			}

		} else {
			buffer = append(buffer, characterBuffer[0])
		}

	}
}

// Reads Pair value as string
func ReadPairValueString(reader io.Reader) (string, error) {
	var buffer []byte = make([]byte, 0, 0)
	var escape bool = false

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return "", errRead
		}

		if escape {
			if characterBuffer[0] == escapeCharacter {
				buffer = append(buffer, characterBuffer[0])
				escape = false
			} else {
				if isControl(characterBuffer[0]) {
					buffer = append(buffer, characterBuffer[0])
					escape = false
				} else {
					return "", errors.New("control byte was expected")
				}
			}
		} else {
			if characterBuffer[0] == escapeCharacter {
				escape = true
			} else {
				if characterBuffer[0] == pairDelimiter {
					str := fmt.Sprintf("%s", buffer)
					return str, nil
				} else {
					if isControl(characterBuffer[0]) {
						return "", errors.New("unexpected control byte")
					} else {
						buffer = append(buffer, characterBuffer[0])
					}
				}
			}
		}

	}
}
