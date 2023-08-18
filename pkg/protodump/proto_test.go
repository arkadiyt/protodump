package protodump

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const FIXTURES = "fixtures"

func convertProtoToFileDescriptor(filePath string) (*descriptorpb.FileDescriptorProto, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	filename := "proto.bin"
	err = exec.Command("protoc", fmt.Sprintf("--go_out=%s", dir), fmt.Sprintf("--descriptor_set_out=%s", path.Join(dir, filename)), filePath).Run()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path.Join(dir, filename))
	if err != nil {
		return nil, err
	}

	var descriptor descriptorpb.FileDescriptorSet
	err = proto.Unmarshal(data, &descriptor)
	if err != nil {
		return nil, err
	}
	return descriptor.GetFile()[0], nil
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

			actual, err := NewFromDescriptor(descriptor)
			assert.NoError(t, err)

			assert.Equal(t, string(expected), actual.String())
		})
	}
}
