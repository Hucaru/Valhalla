package npcChat

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func loadScripts() {
	var files []string

	root := "scripts/npc/"

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range files[1:] {
		loadScript(file)
	}

}

func watchFiles() {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					loadScript(event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					npcID, err := strconv.Atoi(strings.Split(filepath.Base(event.Name), ".")[0])

					if err != nil {
						log.Fatal(err)
					}

					removeScripts(uint32(npcID))
				}

			case err := <-watcher.Errors:
				log.Println("Error in npc script file watcher:", err)
			}
		}
	}()

	if err := watcher.Add("scripts/npc/"); err != nil {
		log.Println("NPC Script watcher error:", err)
	}

	<-done
}

func loadScript(file string) {
	content, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	npcID, err := strconv.Atoi(strings.Split(filepath.Base(file), ".")[0])

	if err != nil {
		log.Fatal(err)
	}

	addScript(uint32(npcID), string(content))
}
