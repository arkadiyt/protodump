package scan

import (
	"bytes"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protowire"
)

func ScanFile(path string) ([][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open file: %w", err)
	}
	return Scan(data), nil
}

func consumeBytes(data []byte, position int) int {
	start := position
	for {
		_, _, len := protowire.ConsumeField(data[position:])
		if len < 0 {
			_ = protowire.ParseError(len)
			return position - start

		}
		position += len
	}
}

func Scan(data []byte) [][]byte {
	results := make([][]byte, 0)

	for {
		index := bytes.Index(data, []byte(".proto"))
		if index == -1 {
			break
		}

		// Look for a 0xa byte: 0 0001 010
		//                      ^ no more bytes for varint
		//                           ^ field 1 (file name)
		//                             ^ type 2 (LEN)
		start := bytes.LastIndexByte(data[:index], 0xa)
		if start == -1 {
			// fmt.Printf("Warn: couldn't find start to proto definition at offset:0x%s\n", strconv.FormatInt(int64(index), 16))
			data = data[index+1:]
			continue
		}

		length := consumeBytes(data, start)
		results = append(results, data[start:start+length])
		data = data[start+length:]
	}

	return results
}
