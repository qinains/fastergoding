/*
Package fastergoding provides a function to automatically compile and run go code.
Example:

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/qinains/fastergoding"
)

func main() {
	fastergoding.Run() // Just add this code

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello %s!", r.URL.Query().Get("name"))
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

*/
package fastergoding

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	mainCmd      *exec.Cmd
	fileModTimes = make(map[string]int64)
	lock         sync.Mutex
)

const runMode = "__RUN_MOD_RELOAD__"

func runCmd(name string, args ...string) {
	cmdStr := name
	for _, arg := range args {
		cmdStr += " " + arg
	}
	log.Printf("Run cmd: %s", cmdStr)
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = append(os.Environ(), "GOGC=off")
	cmd.Run()
}

func restart(rootPath string) {
	lock.Lock()
	defer lock.Unlock()

	runCmd("go", "install")
	runCmd("go", "build")

	defer func() {
		if e := recover(); e != nil {
			log.Printf("Kill recover: %s", e)
		}
	}()
	if mainCmd != nil && mainCmd.Process != nil {
		err := mainCmd.Process.Kill()
		if err != nil {
			log.Printf("Process kill error: %s", err)
		}
	}
	go func() {
		appName := "./" + path.Base(rootPath)
		if runtime.GOOS == "windows" {
			appName += ".exe"
		}
		mainCmd = exec.Command(appName)
		mainCmd.Stdout = os.Stdout
		mainCmd.Stderr = os.Stderr
		mainCmd.Args = append([]string{appName})
		mainCmd.Env = append(os.Environ(), runMode+"="+runMode)
		mainCmd.Run()
	}()
}

func watch(rootPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("New Watcher error: %s", err)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				filePath := event.Name
				if !strings.HasSuffix(filePath, ".go") || strings.Contains(filePath, ".#") {
					continue
				}
				fi, _ := os.Stat(filePath)
				mt := fi.ModTime().Unix()
				mt2 := fileModTimes[filePath]
				fileModTimes[filePath] = mt
				if mt != mt2 {
					go func() {
						log.Printf("Fired by: %s", filePath)
						restart(rootPath)
					}()
				}
			case err := <-watcher.Errors:
				log.Printf("Watcher errors: %#v", err)
			}
		}
	}()

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Fatalf("Watcher add error: %s", err)
			}
			return nil
		}
		return nil
	})
}

/*
Run automatically compile and run the main function when the files is changed.
*/
func Run() {
	if os.Getenv(runMode) == runMode {
		return
	}
	rootPath, _ := os.Getwd()
	os.Chdir(rootPath)

	watch(rootPath)
	restart(rootPath)
	for {
		runtime.Goexit()
	}
}
