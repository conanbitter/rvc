package main

import "regexp"

var regexRange = regexp.MustCompile(`^(.*?)\{\s*(\d+),\s*(\d+)\s*-\s*(\d+)\s*\}(.*)$`)
var regexIndex = regexp.MustCompile(`^(.*?)%(\d+)%(.*)$`)

// Formats:
// - folder                 (all image files: *.png, *.jpg, *.jpeg, *.tif, *.tiff)
// - file1, file2, file3    (all listed existing files)
// - filename%5%.ext        (all files that match pattern, sorted by index)
// - filename{5,1-205}.ext  (all files that match pattern, sorted by index)

func listFiles(input string) []string {
	result := make([]string, 0)
	return result
}
