package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"swamd"

	flag "github.com/spf13/pflag"
)

func main() {
	targetPath := "."
	outputFile := "api_spec.md"

	flag.StringVarP(&targetPath, "p", "p", targetPath, "Target path to parse Go files from (default: .)")
	flag.StringVarP(&outputFile, "o", "o", outputFile,
		fmt.Sprintf("Output file to write API specifications to (default: %s)", outputFile))
	flag.Parse()

	if _, err := os.Stat(outputFile); err == nil {
		err = os.Remove(outputFile)
		if err != nil {
			fmt.Printf("Error removing output file %s: %s\n", outputFile, err)
			return
		}
	}

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			swamd.ParseGoFile(path, outputFile)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
}
