package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/arkadiyt/protodump/pkg/protodump"
)

var debug bool

func Debug(str string, a ...any) (int, error) {
	if debug {
		return fmt.Printf(str, a...)
	}
	return 0, nil
}

func writeFile(outputDir string, filename string, content []byte) (string, error) {
	outputDirAbs, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("couldn't get absolute dir for %s: %v", outputDir, err)
	}

	fileDir, fileBase := filepath.Split(filename)

	parts := strings.Split(path.Clean(fileDir), string(filepath.Separator))
	var i int
	for i = 0; i < len(parts); i++ {
		_, err := os.Stat(filepath.Join(outputDirAbs, filepath.Join(parts[:i+1]...)))
		if os.IsNotExist(err) {
			break
		}
	}

	eval := filepath.Join(outputDirAbs, filepath.Join(parts[:i]...))
	base, err := filepath.EvalSymlinks(eval)
	if err != nil {
		return "", fmt.Errorf("failed to evalsymlinks on %s: %v", eval, err)
	}

	if !strings.HasPrefix(base, outputDirAbs) {
		return "", fmt.Errorf("invalid filepath: %s", base)
	}

	rest := filepath.Join(parts[i:]...)
	err = os.MkdirAll(filepath.Join(base, rest), 0700)
	if err != nil {
		return "", fmt.Errorf("failed to mkdirall on %s: %v", rest, err)
	}

	final := filepath.Join(base, rest, fileBase)
	err = os.WriteFile(final, content, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %v", final, err)
	}
	return final, nil
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Couldn't determine current working directory: %v\n", err)
	}

	var file = flag.String("file", "", "The file to extract definitions from")
	var output = flag.String("output", cwd, "The output directory to save definitions in (will be created if it doesn't exist). Defaults to current directory.")
	flag.BoolVar(&debug, "v", false, "Verbose output")
	flag.Parse()

	if *file == "" {
		fmt.Printf("Usage:\n")
		flag.PrintDefaults()
		return
	}

	results, err := protodump.ScanFile(*file)
	if err != nil {
		log.Fatalf("Got error scanning: %v\n", err)
	}

	err = os.MkdirAll(*output, 0700)
	if err != nil {
		log.Fatalf("Failed to create output folder %s: %v\n", *output, err)
	}

	for _, result := range results {
		definition, err := protodump.NewFromBytes(result)
		if err != nil {
			Debug("Got error parsing definition: %v\n", err)
		} else {

			filename := definition.Filename()
			if strings.HasSuffix(filename, ".proto") {
				final, err := writeFile(*output, filename, []byte(definition.String()))
				if err != nil {
					fmt.Printf("Failed to write %s: %v\n", final, err)
				} else {
					fmt.Printf("Wrote %s\n", final)
				}
			} else {
				// Need to investigate further
			}
		}
	}
}
