package run

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/xiaoenai/tp-micro/cmd/micro/info"
	"github.com/xiaoenai/tp-micro/cmd/micro/run/fsnotify"
)

// RunProject runs project.
func RunProject(newWatchExts []string, newNotWatch []string) {
	if err := os.Chdir(info.AbsPath()); err != nil {
		tp.Fatalf("[micro] Jump working directory failed: %v", err)
	}

	if len(newWatchExts) > 0 {
		watchExts = append(newWatchExts, ".go")
	}
	if len(newNotWatch) > 0 {
		notWatch = append(notWatch, newNotWatch...)
	}

	go rewatch()

	tp.Printf("[micro] Initializing watcher...")

	select {}
}

// getFileModTime retuens unix timestamp of `os.File.ModTime` by given path.
func getFileModTime(path string) int64 {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		tp.Errorf("[micro] Fail to open file[ %s ]", err)
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		tp.Errorf("[micro] Fail to get file information[ %s ]", err)
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

// checkTMPFile returns true if the event was for TMP files.
func checkTMPFile(name string) bool {
	if strings.HasSuffix(strings.ToLower(name), ".tmp") {
		return true
	}
	return false
}

var (
	eventTime = make(map[string]int64)
)

func rewatch() {
	go rerun()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		tp.Errorf("[micro] Fail to create new watcher[ %v ]", err)
		os.Exit(2)
	}
	for _, path := range readAppDirectories("./") {
		tp.Printf("[micro] Watching directory[ %s ]", path)
		err = watcher.Watch(path)
		if err != nil {
			tp.Errorf("[micro] Fail to watch curpathectory[ %s ]", err)
			os.Exit(2)
		}
	}

	for {
		select {
		case e := <-watcher.Event:
			isbuild := true

			// Skip TMP files for Sublime Text.
			if checkTMPFile(e.Name) {
				continue
			}
			if !checkIfWatchExt(e.Name) {
				continue
			}

			mt := getFileModTime(e.Name)
			if t := eventTime[e.Name]; mt == t {
				isbuild = false
			}

			eventTime[e.Name] = mt

			if isbuild {
				tp.Printf("%s", e.String())
				watcher.Close()
				go rewatch()
				return
			}
		case err := <-watcher.Error:
			tp.Warnf("[micro] %s", err.Error()) // No need to exit here
		}
	}
}

var (
	state        sync.Mutex
	cmd          *exec.Cmd
	isFirstStart = true
)

func rerun() {
	state.Lock()
	defer state.Unlock()
	tp.Printf("[micro] Start build...")
	buildCom := exec.Command("go", "build", "-o", info.FileName())
	buildCom.Env = []string{"GOPATH=" + info.Gopath()}
	for _, env := range os.Environ() {
		if strings.HasPrefix(strings.TrimSpace(env), "GOPATH=") {
			continue
		}
		buildCom.Env = append(buildCom.Env, env)
	}
	buildCom.Stdout = os.Stdout
	buildCom.Stderr = os.Stderr
	err := buildCom.Run()
	if err != nil {
		tp.Errorf("[micro] ============== Build failed ===================")
		return
	}
	tp.Printf("[micro] Build was successful")

	var start string
	if isFirstStart {
		isFirstStart = false
		tp.Printf("[micro] Starting app: %s", info.ProjName())
		start = "Start"
	} else {
		tp.Printf("[micro] Restarting app: %s", info.ProjName())
		defer func() {
			if e := recover(); e != nil {
				tp.Printf("[micro] Kill.recover -> %v", e)
			}
		}()
		if cmd != nil && cmd.Process != nil {
			err := cmd.Process.Kill()
			cmd.Process.Release()
			if err != nil {
				tp.Printf("[micro] Kill -> %v", err)
			}
		}
		start = "Restart"
	}

	go func() {
		cmd = exec.Command("./" + info.FileName())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		if err := cmd.Start(); err != nil {
			tp.Errorf("[micro] Fail to start app[ %s ]", err)
			return
		}
		tp.Printf("[micro] %s was successful", start)
		cmd.Wait()
		tp.Printf("[micro] Old process was stopped")
	}()
}

func readAppDirectories(dir string) (paths []string) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		tp.Fatalf("[micro] read project directorys failed: %v", err)
		return
	}
	useDirectory := false
	for _, fileInfo := range fileInfos {
		if checkIfNotWatch(fileInfo.Name()) {
			continue
		}
		if fileInfo.IsDir() == true && fileInfo.Name()[0] != '.' {
			paths = append(paths, readAppDirectories(path.Join(dir, fileInfo.Name()))...)
			continue
		}

		if useDirectory == true {
			continue
		}

		if checkIfWatchExt(fileInfo.Name()) {
			paths = append(paths, dir)
			useDirectory = true
		}
	}
	return
}

var notWatch = []string{}

func checkIfNotWatch(name string) bool {
	if name[0] == '_' {
		return true
	}
	for _, s := range notWatch {
		if name == s {
			return true
		}
	}
	return false
}

var watchExts = []string{".go", ".ini", ".yaml", ".toml", ".xml"}

// checkIfWatchExt returns true if the name HasSuffix <watch_ext>.
func checkIfWatchExt(name string) bool {
	for _, s := range watchExts {
		if strings.HasSuffix(strings.ToLower(name), s) {
			return true
		}
	}
	return false
}
