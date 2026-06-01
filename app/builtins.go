package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

const HOME_PATH = "~"

func echoCommand(args []string) (output, errorMessage string) {
	return strings.Join(args, " "), ""
}

func typeCommand(args []string) (output, errorMessage string) {
	arg := strings.Join(args, " ")

	if _, exists := builtins[arg]; exists {
		return fmt.Sprintf("%s is a shell builtin", arg), ""
	} else if path, exists := pathExecutables[arg]; exists {
		return fmt.Sprintf("%s is %s", arg, path), ""
	} else {
		return "", fmt.Sprintf("%s: not found", arg)
	}
}

func pwdCommand() (output, errorMessage string) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Sprintf("Error getting current directory: %v", err)
	}

	return dir, ""
}

func cdCommand(args []string) (output, errorMessage string) {
	path := strings.Join(args, " ")
	if path == HOME_PATH {
		homeDir, _ := os.UserHomeDir()
		path = homeDir
	}

	err := os.Chdir(path)
	if err != nil {
		return "", fmt.Sprintf("cd: %s: No such file or directory", path)
	}

	return "", ""
}

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
