package protodump

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const FIXTURES = "fixtures"

var regex = regexp.MustCompile("(?m)var .* = \\[\\]byte{([^}]*)}")

func convertProtoToFileDescriptor(filePath string) ([]byte, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	err = exec.Command("protoc", fmt.Sprintf("--go_out=%s", dir), filePath).Run()
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("%s.pb.go", strings.TrimSuffix(path.Base(filePath), ".proto"))
	data, err := os.ReadFile(path.Join(dir, fileName))
	if err != nil {
		return nil, err
	}

	match := regex.FindSubmatch(data)
	if len(match) != 2 {
		return nil, fmt.Errorf("Could not find file descriptor bytes in %s", fileName)
	}

	var hexString strings.Builder
	for _, char := range strings.ToLower(string(bytes.ReplaceAll(match[1], []byte("0x"), []byte{}))) {
		r := rune(char)
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') {
			hexString.WriteRune(r)
		}
	}

	result, err := hex.DecodeString(hexString.String())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func TestDefinitions(t *testing.T) {
	files, err := os.ReadDir(FIXTURES)
	assert.NoError(t, err)

	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
			filePath := path.Join(FIXTURES, file.Name())
			descriptor, err := convertProtoToFileDescriptor(filePath)
			assert.NoError(t, err)

			expected, err := os.ReadFile(filePath)
			assert.NoError(t, err)

			actual, err := NewFromBytes(descriptor)
			assert.NoError(t, err)

			assert.Equal(t, string(expected), actual.String())
		})
	}
}
