package tpl

import (
	"log"
	"os"
	"path"
)

func GetByFilename(filename string) string {
	fullpath := path.Join("templates", filename)

	bytes, err := os.ReadFile(fullpath)
	if err != nil {
		log.Fatalf("Failed to read template content of %v: %v", filename, err)
	}
	return string(bytes)
}
