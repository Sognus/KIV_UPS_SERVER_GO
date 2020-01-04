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

	// Wait limits before decode error
	limitStart = 128
	limitHeader = 64
	limitInt = 20
	limitString = 128

	// Error threshold
	errorThreshold = 0.66
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

func Decoder(serverContext *server, client *Client) {
	buffer := bufio.NewReader(client.Reader)

	var total int = 0
	var good int = 0

	for {
		errDecode := Decode(buffer)

		if errDecode != nil {
			total++
		} else {
			good++
			total++
		}

		bad := total - good
		var percent float64 = float64(bad / total)

		if percent > errorThreshold && total > 16  {
			_ = RemoveClient(serverContext, client.Socket)
			break
		}

	}
}

func Decode(buffer *bufio.Reader) error {
	// Declare values
	var errReadHeader error
	var errReadValue error
	var valueInt int

	var msg message
	errStart := WaitForStart(buffer)

	// Ignore error
	if errStart != nil {
		return errStart
	}

	// Read bytes until  "id:" is read
	_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "id")

	if errReadHeader != nil {
		return errReadHeader
	}

	// Read bytes for integer value
	valueInt, errReadValue = ReadPairValueInt(buffer)

	if errReadValue != nil {
		return errReadValue
	}

	// Set message ID
	msg.id = valueInt

	// Read bytes until  "rid:" is read
	_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "rid")

	if errReadHeader != nil {
		return errReadHeader
	}

	// Read bytes for integer value
	valueInt, errReadValue = ReadPairValueInt(buffer)

	if errReadValue != nil {
		return errReadValue
	}

	// Set message return ID
	msg.rid = valueInt

	// Read bytes until "type:" is read
	_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "type")

	if errReadHeader != nil {
		return errReadHeader
	}

	// Read bytes for integer value
	valueInt, errReadValue = ReadPairValueInt(buffer)

	if errReadValue != nil {
		return errReadValue
	}

	// Set message type
	msg.msg = valueInt

	// Expect header end byte
	errExpect := ExpectByte(buffer, headEnd)

	if errExpect != nil {
		return errExpect
	}

	// Read message content
	msg.content = make(map[string]string)


	for {
		headerData, errReadHeader := ReadPairHeaderAny(buffer, valueDelimiter)

		if errReadHeader != nil {
			return errReadHeader
		}

		fmt.Printf("PairHeader: %s\n", headerData)

		valueData, errReadValue := ReadPairValueString(buffer)

		if errReadValue != nil {
			return errReadValue
		}

		fmt.Printf("PairValue: %s\n", valueData)

		fmt.Printf("msg.content[%s] = %s\n", headerData, valueData)
		msg.content[string(headerData)] = valueData

		errPeek := PeekExpectByte(buffer, endCharacter)


		if errPeek != nil {
			continue
		} else {
			return nil
		}
	}
}

// Reads bytes until start character is read
func WaitForStart(reader io.Reader) error {
	var limit int = limitStart

	for {
		startBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, startBuffer)

		if errRead != nil {
			return errRead
		}

		if startBuffer[0] == 0 {
			continue
		}

		if startBuffer[0] == startCharacter {
			return nil
		} else {
			limit--
		}

		if limit == 0 {
			return errors.New("waitForStart: limit exceeded")
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
	var limit int = limitHeader

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return nil, errRead
		}

		if limit == 0 {
			return nil, errors.New("readPairHeaderAny: limit exceeded")
		}

		if escape {
			if characterBuffer[0] == escapeCharacter {
				buffer = append(buffer, characterBuffer[0])
				escape = false
				limit--
			} else {
				if isControl(characterBuffer[0]) {
					buffer = append(buffer, characterBuffer[0])
					escape = false
					limit--
				} else {
					return nil, errors.New("control byte was expected")
				}
			}
		} else {
			if characterBuffer[0] == escapeCharacter {
				escape = true
				limit--
			} else {
				if characterBuffer[0] == character {
					return buffer, nil
				} else {
					if isControl(characterBuffer[0]) {
						return nil, errors.New("unexpected control byte")
					} else {
						buffer = append(buffer, characterBuffer[0])
						limit--
					}
				}
			}
		}
	}
}

// Reads bytes and detects given header in message - escape not accepted
func ReadPairHeader(reader io.Reader, character byte, expected string) ([]byte, error) {
	var buffer []byte = make([]byte, 0, 0)
	var limit int = limitHeader

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return nil, errRead
		}

		if limit == 0 {
			return nil, errors.New("readPairHeader: limit exceeded")
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
			limit--
		}
	}
}

func ReadPairValueInt(reader io.Reader) (int, error) {
	var buffer []byte = make([]byte, 0, 0)
	var limit int = limitInt

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return 0, errRead
		}

		if limit == 0 {
			return 0, errors.New("readPairValueInt: limit exceeded")
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
			limit--
		}

	}
}

// Reads Pair value as string
func ReadPairValueString(reader io.Reader) (string, error) {
	var buffer []byte = make([]byte, 0, 0)
	var escape bool = false
	var limit int = limitString

	for {
		characterBuffer := make([]byte, 1)
		_, errRead := io.ReadFull(reader, characterBuffer)

		if errRead != nil {
			return "", errRead
		}
		if limit == 0 {
			return "", errors.New("readPairValueString: limit exceeded")
		}

		if escape {
			if characterBuffer[0] == escapeCharacter {
				buffer = append(buffer, characterBuffer[0])
				escape = false
				limit--
			} else {
				if isControl(characterBuffer[0]) {
					buffer = append(buffer, characterBuffer[0])
					escape = false
					limit--
				} else {
					return "", errors.New("control byte was expected")
				}
			}
		} else {
			if characterBuffer[0] == escapeCharacter {
				escape = true
				limit--
			} else {
				if characterBuffer[0] == pairDelimiter {
					str := fmt.Sprintf("%s", buffer)
					return str, nil
				} else {
					if isControl(characterBuffer[0]) {
						return "", errors.New("unexpected control byte")
					} else {
						buffer = append(buffer, characterBuffer[0])
						limit--
					}
				}
			}
		}

	}
}
