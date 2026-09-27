package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

var protoContent = []byte("syntax = \"proto3\";\n")

func testOutput(t *testing.T) *os.Root {
	t.Helper()
	root, err := openOutputDir(filepath.Join(t.TempDir(), "output"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	return root
}

func TestWriteFileCreatesProto(t *testing.T) {
	root := testOutput(t)
	for _, name := range []string{"service.proto", "api/v1/service.proto", "with spaces/service.proto"} {
		t.Run(name, func(t *testing.T) {
			filename, err := writeFile(root, name, protoContent)
			require.NoError(t, err)
			require.Equal(t, filepath.Join(root.Name(), filepath.FromSlash(name)), filename)
			content, err := os.ReadFile(filename)
			require.NoError(t, err)
			require.Equal(t, protoContent, content)
		})
	}
}

func TestWriteFileRejectsInvalidNamesBeforeCreatingDirectories(t *testing.T) {
	for _, name := range []string{
		"", ".", "..", "../outside.proto", "/absolute.proto",
		"api/../service.proto", "api/./service.proto", "api//service.proto",
		`api\service.proto`, `C:/service.proto`, `C:service.proto`,
		`//server/share/service.proto`, `\\server\share\service.proto`,
		"api/stream:service.proto", "api/service.txt", "api/service.proto/",
		"api/\x00.proto", "api/\xff.proto",
	} {
		t.Run(name, func(t *testing.T) {
			root := testOutput(t)
			_, err := writeFile(root, name, protoContent)
			require.Error(t, err)
			entries, err := os.ReadDir(root.Name())
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}

func TestWriteFileRejectsWindowsDevices(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows device names are platform-specific")
	}
	for _, name := range []string{"NUL/service.proto", "CON/service.proto", "COM1/service.proto"} {
		t.Run(name, func(t *testing.T) {
			root := testOutput(t)
			_, err := writeFile(root, name, protoContent)
			require.Error(t, err)
			entries, err := os.ReadDir(root.Name())
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}

func TestWriteFilePreservesExistingFiles(t *testing.T) {
	root := testOutput(t)
	original := []byte("original contents")
	filename := filepath.Join(root.Name(), "existing.proto")
	require.NoError(t, os.WriteFile(filename, original, 0600))

	_, err := writeFile(root, "existing.proto", protoContent)
	require.ErrorIs(t, err, os.ErrExist)
	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, original, content)
}

func TestWriteFileRejectsExternalDirectoryLinks(t *testing.T) {
	for _, relative := range []bool{false, true} {
		name := "absolute"
		if relative {
			name = "relative"
		}
		t.Run(name, func(t *testing.T) {
			root := testOutput(t)
			outside := t.TempDir()
			target := outside
			if relative {
				var err error
				target, err = filepath.Rel(root.Name(), outside)
				require.NoError(t, err)
			}
			require.NoError(t, os.Symlink(target, filepath.Join(root.Name(), "linked")))

			_, err := writeFile(root, "linked/new/service.proto", protoContent)
			require.Error(t, err)
			entries, err := os.ReadDir(outside)
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}

func TestWriteFilePreservesFileLinks(t *testing.T) {
	for _, kind := range []string{"symlink", "dangling symlink", "hard link"} {
		t.Run(kind, func(t *testing.T) {
			root := testOutput(t)
			target := filepath.Join(t.TempDir(), "target.proto")
			original := []byte("original contents")
			if kind != "dangling symlink" {
				require.NoError(t, os.WriteFile(target, original, 0600))
			}
			link := filepath.Join(root.Name(), "linked.proto")
			if kind == "hard link" {
				require.NoError(t, os.Link(target, link))
			} else {
				require.NoError(t, os.Symlink(target, link))
			}

			_, err := writeFile(root, "linked.proto", protoContent)
			require.Error(t, err)
			content, err := os.ReadFile(target)
			if kind == "dangling symlink" {
				require.ErrorIs(t, err, os.ErrNotExist)
			} else {
				require.NoError(t, err)
				require.Equal(t, original, content)
			}
		})
	}
}

func TestWriteFileAllowsInternalDirectoryLinks(t *testing.T) {
	root := testOutput(t)
	require.NoError(t, root.Mkdir("schemas", 0700))
	require.NoError(t, os.Symlink("schemas", filepath.Join(root.Name(), "linked")))

	_, err := writeFile(root, "linked/v1/service.proto", protoContent)
	require.NoError(t, err)
	content, err := root.ReadFile(filepath.FromSlash("schemas/v1/service.proto"))
	require.NoError(t, err)
	require.Equal(t, protoContent, content)
}

func TestOpenOutputDirAllowsUserSelectedSymlink(t *testing.T) {
	target := t.TempDir()
	link := filepath.Join(t.TempDir(), "output")
	require.NoError(t, os.Symlink(target, link))
	root, err := openOutputDir(link)
	require.NoError(t, err)
	defer root.Close()

	_, err = writeFile(root, "service.proto", protoContent)
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(target, "service.proto"))
	require.NoError(t, err)
	require.Equal(t, protoContent, content)
}

func TestOutputRootRetainsDirectoryHandle(t *testing.T) {
	root := testOutput(t)
	original := root.Name()
	moved := original + "-moved"
	err := os.Rename(original, moved)
	if runtime.GOOS == "windows" {
		// Windows prevents renaming a directory while its root handle is open.
		require.Error(t, err)
		moved = original
	} else {
		require.NoError(t, err)
		require.NoError(t, os.Mkdir(original, 0700))
	}

	_, err = writeFile(root, "service.proto", protoContent)
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Join(moved, "service.proto"))
	require.NoError(t, err)
	require.Equal(t, protoContent, content)
	if moved != original {
		entries, err := os.ReadDir(original)
		require.NoError(t, err)
		require.Empty(t, entries)
	}
}
