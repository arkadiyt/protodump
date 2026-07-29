package protodump

import (
	"bytes"
	"fmt"
	"os"
	"maps"
	"slices"

	"google.golang.org/protobuf/encoding/protowire"
)

const scan = ".proto"
var tags = map[string]uint64{
	"name":              0xa,
	"package":           0x12,
	"dependency":        0x1a,
	"public_dependency": 0x50,
	"weak_dependency":   0x58,
	"message_type":      0x22,
	"enum_type":         0x2a,
	"service":           0x32,
	"extension":         0x3a,
	"options":           0x42,
	"source_code_info":  0x4a,
	"syntax":            0x62,
	"edition":           0x70,
}

func consumeBytes(data []byte, position int, requiredFields int) (int, error) {
	start := position
	validTags := slices.Collect(maps.Values(tags))
	seenName := false
	seenFields := 0
	
	for {
		if position >= len(data) {
			goto EndOfData
		}

		tag, n := protowire.ConsumeVarint(data[position:])
		// Treat "invalid field number" as end of data, not an error
		if n <= 0 || !slices.Contains(validTags, tag) {
			goto EndOfData
		}
		if tag == tags["name"] {
			if seenName {
				// Only consume Field 1 once (to handle the case where protobuf definitions are adjacent in program memory)
				goto EndOfData
			}
			seenName = true
		}

		fieldNum, _, length := protowire.ConsumeField(data[position:])
		if length < 0 {
			err := protowire.ParseError(length)
			// Return other parse errors as actual errors
			return position - start, fmt.Errorf("couldn't consume proto bytes: %w", err)
		}

		seenFields += 1

		if fieldNum == 1 {
			seenName = true
		}

		// Prevent infinite loop - if we can't consume any bytes, we're done
		if length == 0 {
			goto EndOfData
		}

		position += length

		// Additional safety check - don't read beyond data bounds
		if position-start >= len(data[start:]) {
			goto EndOfData
		}
	}

EndOfData:
	err := func() error {
		if seenFields < requiredFields {
			return fmt.Errorf("couldn't consume proto bytes: incomplete proto definition")
		}
		return nil
	}()
	return position - start, err
}

func ScanFile(path string, requiredFields int) ([][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %w", err)
	}
	return Scan(data, requiredFields), nil
}

func Scan(data []byte, requiredFields int) [][]byte {
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
		start := bytes.LastIndexByte(data[:index], byte(tags["name"]))
		if start == -1 {
			data = data[index+1:]
			continue
		}

		// If the "<file>.proto" filename is 10 characters long then the 0xa byte we found is actually
		// the string length and not the start of the protobuf record, so go back 1 more
		if index-start == (int(tags["name"])-len(scan)+1) && start > 0 && data[start-1] == byte(tags["name"]) {
			start -= 1
		}

		length, err := consumeBytes(data, start, requiredFields)
		if err != nil {
			fmt.Printf("%v\n", err)
			if len(data) > index {
				data = data[index+1:]
				continue
			} else {
				break
			}
		}
		results = append(results, data[start:start+length])
		data = data[start+length:]
	}

	return results
}
