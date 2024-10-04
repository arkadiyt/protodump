package protodump

import (
	"bytes"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protowire"
)

const scan = ".proto"
const magicByte = 0xa

func consumeBytes(data []byte, position int) int {
	start := position
	consumedFieldOne := false
	for {
		number, _, len := protowire.ConsumeField(data[position:])
		if len < 0 {
			_ = protowire.ParseError(len)
			return position - start

		}

		// Only consume Field 1 once (to handle the case where protobuf definitions are adjacent
		// in program memory)
		if number == 1 {
			if consumedFieldOne {
				return position - start
			}
			consumedFieldOne = true
		}

		position += len
	}
}

func ScanFile(path string) ([][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open file: %w", err)
	}
	return Scan(data), nil
}

func Scan(data []byte) [][]byte {
	results := make([][]byte, 0)

	for {
		index := bytes.Index(data, []byte(scan))
		if index == -1 {
			break
		}

		// Look for a 0xa byte: 0 0001 010
		//                      ^ no more bytes for varint
		//                        ^ field 1 (file name)
		//                             ^ type 2 (LEN)
		start := bytes.LastIndexByte(data[:index], magicByte)
		if start == -1 {
			data = data[index+1:]
			continue
		}

		// If the "<file>.proto" filename is 10 characters long then the 0xa byte we found is actually
		// the string length and not the start of the protobuf record, so go back 1 more
		if index-start == (magicByte-len(scan)+1) && start > 0 && data[start-1] == magicByte {
			start -= 1
		}

		length := consumeBytes(data, start)
		if start == 0 && length == 0 {
			data = data[index+1:]
			continue
		}
		results = append(results, data[start:start+length])
		data = data[start+length:]
	}

	return results
}
