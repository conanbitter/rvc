package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var regexRange = regexp.MustCompile(`^(.*?)\{\s*(\d+),\s*(\d+)\s*-\s*(\d+)\s*\}(.*)$`)
var regexIndex = regexp.MustCompile(`^(.*?)%(\d+)%(.*)$`)
var regexExt = regexp.MustCompile(`\.(tiff|tif|png|jpg|jpeg)$`)

type IndexedFilename struct {
	Filename string
	Index    int
}

// Formats:
// - folder                 (all image files: *.png, *.jpg, *.jpeg, *.tif, *.tiff)
// - file1, file2, file3    (all listed existing files)
// - filename%5%.ext        (all files that match pattern, sorted by index)
// - filename{5,1-205}.ext  (all files that match pattern, sorted by index)

func addAbsFile(filelist []string, newFile string) []string {
	absFile, _ := filepath.Abs(newFile)
	return append(filelist, absFile)
}

func listFiles(input string) []string {
	result := make([]string, 0)

	// Comma type
	if strings.Contains(input, ",") && (!regexRange.MatchString(input)) {
		files := strings.Split(input, ",")
		for _, file := range files {
			info, err := os.Stat(file)
			if !os.IsNotExist(err) && !info.IsDir() {
				result = addAbsFile(result, strings.TrimSpace(file))
			}
		}
		return result
	}

	// Folder type
	fileInfo, err := os.Stat(input)
	if err == nil && fileInfo.IsDir() {
		entries, err := os.ReadDir(input)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && regexExt.MatchString(entry.Name()) {
					result = addAbsFile(result, filepath.Join(input, entry.Name()))
				}
			}
			return result
		}
	}

	// Pattern types
	fpath := filepath.Dir(input)
	fname := filepath.Base(input)
	var (
		digits    int
		start     int
		end       int
		leftPart  string
		rightPart string
	)

	groups := regexIndex.FindStringSubmatch(fname)
	if len(groups) != 0 {
		leftPart = groups[1]
		rightPart = groups[3]
		digits, _ = strconv.Atoi(groups[2])
		start = 0
		end = -1
	} else {
		groups = regexRange.FindStringSubmatch(fname)
		if len(groups) == 0 {
			//single file
			result = addAbsFile(result, input)
			return result
		}
		leftPart = groups[1]
		rightPart = groups[5]
		digits, _ = strconv.Atoi(groups[2])
		start, _ = strconv.Atoi(groups[3])
		end, _ = strconv.Atoi(groups[4])
	}

	matchRegexp := regexp.MustCompile(fmt.Sprintf("^%s(\\d{%d})%s$", regexp.QuoteMeta(leftPart), digits, regexp.QuoteMeta(rightPart)))
	entries, err := os.ReadDir(fpath)
	if err == nil {
		indexedFiles := make([]IndexedFilename, 0)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			groups = matchRegexp.FindStringSubmatch(entry.Name())
			if len(groups) == 0 {
				continue
			}
			index, _ := strconv.Atoi(groups[1])
			if index < start {
				continue
			}
			if end > 0 && index > end {
				continue
			}
			indexedFiles = append(indexedFiles, IndexedFilename{Filename: entry.Name(), Index: index})
		}
		sort.Slice(indexedFiles, func(i, j int) bool { return indexedFiles[i].Index < indexedFiles[j].Index })
		for _, ifentry := range indexedFiles {
			result = addAbsFile(result, filepath.Join(fpath, ifentry.Filename))
		}
		return result
	}

	return result
}
