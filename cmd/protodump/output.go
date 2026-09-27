package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func openOutputDir(dir string) (*os.Root, error) {
	// These platforms do not provide handle-based, race-resistant roots.
	if runtime.GOOS == "js" || runtime.GOOS == "plan9" {
		return nil, fmt.Errorf("secure output is not supported on %s", runtime.GOOS)
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	return os.OpenRoot(dir)
}

func writeFile(output *os.Root, filename string, content []byte) (string, error) {
	// Descriptor names must be slash-separated relative paths on every platform.
	if strings.ContainsAny(filename, `\:`) || !strings.HasSuffix(filename, ".proto") {
		return "", fmt.Errorf("invalid proto filename: %q", filename)
	}
	name, err := filepath.Localize(filename)
	if err != nil {
		return "", fmt.Errorf("invalid proto filename %q: %w", filename, err)
	}

	// Both operations enforce containment, even if symlinks change between them.
	if err := output.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return "", fmt.Errorf("couldn't create directories for %q: %w", filename, err)
	}
	file, err := output.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", fmt.Errorf("couldn't create %q: %w", filename, err)
	}

	_, writeErr := file.Write(content)
	closeErr := file.Close()
	if writeErr != nil {
		return "", fmt.Errorf("couldn't write %q: %w", filename, writeErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("couldn't close %q: %w", filename, closeErr)
	}
	return filepath.Join(output.Name(), name), nil
}
