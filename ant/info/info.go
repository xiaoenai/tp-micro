package info

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// projInfo current project information
var projInfo = struct {
	projName, fileName, absPath, projPath, gopath, rawCmdPath string
}{}

// Init initializes project information.
func Init(appPath string) (err error) {
	if len(appPath) == 0 {
		return errors.New("project(<app_path, a>) path can not be empty!")
	}
	projInfo.absPath = strings.TrimSpace(appPath)
	projInfo.absPath, err = filepath.Abs(projInfo.absPath)
	if err != nil {
		return errors.New("get absolute project path failed: %s!" + err.Error())
	}
	const SRC = "/src/"
	projInfo.absPath = strings.Replace(projInfo.absPath, `\`, `/`, -1)
	projInfo.absPath = strings.TrimRight(projInfo.absPath, "/")
	srcIdx := strings.Index(projInfo.absPath, SRC)
	if srcIdx == -1 {
		return errors.New("the project directory must be in the $GOPATH/src!")
	}
	projInfo.gopath = projInfo.absPath[:srcIdx]
	projInfo.projPath = projInfo.absPath[srcIdx+len(SRC):]
	projInfo.projName = projInfo.absPath[strings.LastIndex(projInfo.absPath, "/")+1:]
	projInfo.fileName = projInfo.projName
	if runtime.GOOS == "windows" {
		projInfo.fileName += ".exe"
	}
	projInfo.rawCmdPath, _ = os.Getwd()
	return nil
}

// ProjName returns project name.
func ProjName() string {
	return projInfo.projName
}

// FileName returns the binary execution file name.
func FileName() string {
	return projInfo.fileName
}

// AbsPath returns the absolute project path.
func AbsPath() string {
	return projInfo.absPath
}

// ProjPath returns the relative src project path.
func ProjPath() string {
	return projInfo.projPath
}

// Gopath returns the current $GOPATH.
func Gopath() string {
	return projInfo.gopath
}

// RawCmdPath returns the initial path to start.
func RawCmdPath() string {
	return projInfo.rawCmdPath
}
