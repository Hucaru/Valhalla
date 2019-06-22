package script

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type scriptFile struct {
	name     string
	contents string
	remove   bool
}

var fileChan = make(chan scriptFile)

func init() {
	go collector()
}

func collector() {
	for {
		script := <-fileChan

		if script.remove {
			loadedScripts.remove(strings.TrimSuffix(script.name, filepath.Ext(script.name)))
		} else {
			loadedScripts.add(script)
		}
	}
}

func readScript(file string) scriptFile {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	script := scriptFile{name: file, contents: string(data), remove: false}

	return script
}

func loadScripts(directory string) {
	scripts := []scriptFile{}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			scripts = append(scripts, readScript(path))
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, s := range scripts {
		fileChan <- s
	}
}

func WatchScriptDirectory(directory string) {
	loadScripts(directory)

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					done <- true
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					script := readScript(event.Name)
					fileChan <- script
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					script := scriptFile{name: event.Name, remove: true}
					fileChan <- script
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					done <- true
					return
				}

				log.Fatal(err)
			}
		}
	}()

	err = watcher.Add(directory)

	if err != nil {
		log.Fatal(err)
	}

	<-done
}
