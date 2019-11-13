package tpl

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiaoenai/tp-micro/v6/micro/info"
)

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	data = bytes.Replace(data, projNameTpl, projNameBytes, -1)
	data = bytes.Replace(data, projPathTpl, projPathBytes, -1)
	if strings.HasSuffix(name, ".go") {
		data, _ = format.Source(data)
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	fmt.Printf("generate %s\n", string(projPathBytes)+"/"+_filePath(dir, name))
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		if name == "__tp-micro__gen__.lock" {
			return nil
		}
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		if child == "logic" {
			continue
		}
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

var (
	projNameBytes []byte
	projPathBytes []byte
	projNameTpl   = []byte("{{PROJ_NAME}}")
	projPathTpl   = []byte("{{PROJ_PATH}}")
)

// Create creates base files.
func Create() {
	projNameBytes = []byte(info.ProjName())
	projPathBytes = []byte(info.ProjPath())
	RestoreAssets("./", "")
}
