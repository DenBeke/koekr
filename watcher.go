package main

import (
	"os"
	"strings"

	"github.com/howeyc/fsnotify"
	log "github.com/sirupsen/logrus"
)

func (k *Koekr) WatchForChanges() {
	// Watch for changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Handle file change
	go func() {
		for {
			select {
			case event := <-watcher.Event:

				if _, err := os.Stat(event.Name); os.IsNotExist(err) {
					break
				}

				if strings.Contains(event.Name, ".go") {
					// Skip Go files
					break
				}

				log.Println("Detected change to:", event.Name)

				if strings.Contains(event.Name, "assets") {
					// Copy asset file
					err = CopyFile(event.Name, "./generated/"+event.Name)
					if err != nil {
						log.Warnln("Couldn't update asset:", err)
					}
					break
				}
				if strings.Contains(event.Name, "pages") {
					// Regenerate page
					k.generatePage(event.Name)
					break
				}
				if strings.Contains(event.Name, k.config.configFile) {
					// Parse config again and regenerated all pages
					err = k.ParseConfig()
					if err == nil {
						k.GenerateAllPages()
					}
				}
				if strings.Contains(event.Name, k.config.template) {
					// Parse template again en regenerate all files
					err = k.ParseTemplates()
					if err == nil {
						k.GenerateAllPages()
					}
				}

			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch("assets")
	err = watcher.Watch("pages")
	err = watcher.Watch("./")
	if err != nil {
		log.Warnln("Couldn't watch for changes:", err)
	} else {
		log.Println("Watching for changes")
	}

	// Hang so program doesn't exit
	<-done

	watcher.Close()
}
