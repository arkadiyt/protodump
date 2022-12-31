package main

import (
    	"flag"
	"fmt"
	"log"
	"os"
	"path"
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

	for _, result := range results {
		definition, err := protodump.NewFromBytes(result)
		if err != nil {
			Debug("Got error parsing definition: %v\n", err)
		} else {
			filename := definition.Filename()
			if strings.HasSuffix(filename, ".proto") {
				dir := path.Join(*output, path.Dir(filename))
				final := path.Join(dir, path.Base(filename))
				os.MkdirAll(dir, 0700)
				os.WriteFile(final, []byte(definition.String()), 0700)
				fmt.Printf("Wrote %s\n", final)
			} else {
				// Need to investigate further
			}
		}
	}
}
