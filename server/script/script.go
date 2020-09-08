package script

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/fsnotify/fsnotify"
)

// Store container
type Store struct {
	folder   string
	scripts  map[string]*goja.Program
	dispatch chan func()
}

func (s Store) String() string {
	return fmt.Sprintf("%v", s.scripts)
}

// CreateStore for scripts
func CreateStore(folder string, dispatch chan func()) *Store {
	return &Store{folder: folder, dispatch: dispatch, scripts: make(map[string]*goja.Program)}
}

// Get script from store
func (s *Store) Get(name string) (*goja.Program, bool) {
	program, ok := s.scripts[name]
	return program, ok
}

// Monitor the script directory and hot load scripts
func (s *Store) Monitor() {
	err := filepath.Walk(s.folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		s.dispatch <- func() {
			name, program, err := createProgramFromFilename(path)

			if err == nil {
				s.scripts[name] = program
			} else {
				log.Println("Script compiling:", err)
			}
		}

		return nil
	})

	if err != nil {
		return
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Println(err)
	}

	defer watcher.Close()

	err = watcher.Add(s.folder)

	if err != nil {
		log.Println(err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				s.dispatch <- func() {
					log.Println("Script:", event.Name, "modified/created")
					name, program, err := createProgramFromFilename(event.Name)

					if err == nil {
						s.scripts[name] = program
					} else {
						log.Println("Script compiling:", err)
					}
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				s.dispatch <- func() {
					name := filepath.Base(event.Name)
					name = strings.TrimSuffix(name, filepath.Ext(name))

					if _, ok := s.scripts[name]; ok {
						log.Println("Script:", event.Name, "removed")
						delete(s.scripts, name)
					} else {
						log.Println("Script: could not find:", name, "to delete")
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			log.Println(err)
		}
	}
}

func createProgramFromFilename(filename string) (string, *goja.Program, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return "", nil, err
	}

	program, err := goja.Compile(filename, string(data), false)

	if err != nil {
		return "", nil, err
	}

	filename = filepath.Base(filename)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	return name, program, nil
}
