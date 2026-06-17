package main

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// executables maps an executable's command name to its full path on disk,
// built once at startup from the PATH environment variable.
var executables map[string]string

func getPathExecutables(path string) map[string]string {
	pathExecutables := make(map[string]string)

	// take in path environment variable and split it into directories
	paths := strings.SplitSeq(path, string(os.PathListSeparator))
	// for each path, walk through the directory for each file
	for path := range paths {
		if len(path) == 0 {
			continue
		}

		currDir, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		// for each file, check if it is executable and add it to the path executable map
		for _, file := range currDir {
			if file.IsDir() {
				continue
			}

			fullPath := filepath.Join(path, file.Name())
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}

			if runtime.GOOS == "windows" && isExecutable(fullPath) { // windows doesn't use POSIX permission bits, so we have to check the file extension instead
				pathExecutables[strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))] = strings.TrimSuffix(fullPath, filepath.Ext(fullPath))
			} else if info.Mode().Perm()&0111 != 0 { // check if file permissions indicate executable
				pathExecutables[file.Name()] = fullPath
			}
		}
	}

	return pathExecutables
}

func isExecutable(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	pathext := os.Getenv("PATHEXT")
	if pathext == "" {
		pathext = ".COM;.EXE;.BAT;.CMD"
	}

	return slices.Contains(strings.Split(strings.ToLower(pathext), ";"), ext)
}
