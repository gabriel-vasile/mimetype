package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

func main() {
	// Define flags to satisfy the Makefile's 'version' target
	versionFlag := flag.Bool("version", false, "display the version of mimetype-detect")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: mimetype-detect [flags] <file1> [file2...]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		fmt.Println("mimetype-detect version 1.0.0")
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	exitCode := 0
	for _, path := range args {
		mtype, err := mimetype.DetectFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error detecting %s: %v\n", path, err)
			exitCode = 1
			continue
		}

		// Print the path, the MIME type string, and the suggested extension
		fmt.Printf("%s: %s (%s)\n", path, mtype.String(), mtype.Extension())
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
