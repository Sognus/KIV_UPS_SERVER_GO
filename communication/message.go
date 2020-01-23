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
	limitStart = 512
	limitHeader = 64
	limitInt = 32
	limitString = 128

	// Error threshold
	errorThreshold = 0.66666
)

type Message struct {
	// Message ID
	Id int
	// Message return ID
	Rid int
	// Message type
	Msg int
	// Message source (0 = server, >0 = client)
	Source int
	// Message content
	Content map[string]string

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

func Decoder(serverContext *Server, client *Client) {
	buffer := bufio.NewReader(client.Reader)

	var total int = 0
	var good int = 0

	for {
		msg, errDecode := Decode(buffer)

		if errDecode != nil {
			// If client was terminated, stop decoding
			if errDecode == io.ErrClosedPipe {
				_ = RemoveClient(serverContext, client.Socket)
				return
			}
			total++
		} else {
			good++
			total++

			// Fill source
			msg.Source = client.UID
			// Print message for debug
			// fmt.Printf("message: %v\n", msg) 
			serverContext.MessageChannel <- *msg
		}

		var percent float64 = float64(good / total)

		if percent < errorThreshold && total > 16  {
			good = 0
			total = 0
			_ = RemoveClient(serverContext, client.Socket)
			return
		}

	}
}

func Decode(buffer *bufio.Reader) (*Message,error) {
	// Declare values
	var errReadHeader error
	var errReadValue error
	var valueInt int

	var msg Message
	errStart := WaitForStart(buffer)

	// Ignore error
	if errStart != nil {
		return nil,errStart
	}

	// Read bytes until  "id:" is read
	_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "id")

	if errReadHeader != nil {
		return nil,errReadHeader
	}

	// Read bytes for integer value
	valueInt, errReadValue = ReadPairValueInt(buffer)

	if errReadValue != nil {
		return nil, errReadValue
	}

	// Set Message ID
	msg.Id = valueInt

	// Read bytes until  "rid:" is read
	_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "rid")

	if errReadHeader != nil {
		return nil, errReadHeader
	}

	// Read bytes for integer value
	valueInt, errReadValue = ReadPairValueInt(buffer)

	if errReadValue != nil {
		return nil, errReadValue
	}

	// Set Message return ID
	msg.Rid = valueInt

	// Read bytes until "type:" is read
	_, errReadHeader = ReadPairHeader(buffer, valueDelimiter, "type")

	if errReadHeader != nil {
		return nil, errReadHeader
	}

	// Read bytes for integer value
	valueInt, errReadValue = ReadPairValueInt(buffer)

	if errReadValue != nil {
		return nil, errReadValue
	}

	// Set Message type
	msg.Msg = valueInt

	// Expect header end byte
	errExpect := ExpectByte(buffer, headEnd)

	if errExpect != nil {
		return nil, errExpect
	}

	// Read Message content
	msg.Content = make(map[string]string)


	for {
		headerData, errReadHeader := ReadPairHeaderAny(buffer, valueDelimiter)

		if errReadHeader != nil {
			return nil,errReadHeader
		}

		valueData, errReadValue := ReadPairValueString(buffer)

		if errReadValue != nil {
			return nil,errReadValue
		}

		msg.Content[string(headerData)] = valueData

		errPeek := PeekExpectByte(buffer, endCharacter)


		if errPeek != nil {
			continue
		} else {
			return &msg,nil
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

// Reads bytes and detects given header in Message - escape not accepted
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

		// Catch unexpected control character
		if isControl(characterBuffer[0]) && characterBuffer[0] != character {
			return nil, errors.New("unexpected control character")
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
