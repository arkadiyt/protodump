package proto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefinitions(t *testing.T) {
	files, err := os.ReadDir("fixtures")
	assert.NoError(t, err)

	for _, file := range files {
		t.Run(file.Name(), func(t *testing.T) {
		})
	}
}
