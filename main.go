package main

import (
	"github.com/BurntSushi/toml"
	"github.com/howeyc/fsnotify"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Koekr struct {
	variables map[string]interface{}
	t         *template.Template
}

func (k *Koekr) findPages() (files []string) {
	searchDir := "./pages/"

	_ = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {

		fi, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if fi.Mode().IsRegular() {
			files = append(files, path)
		}

		return nil
	})

	return files
}

func (k *Koekr) parsePage(content string) map[string]interface{} {
	// Split page config from page content
	pageConfig := []string{}
	pageContent := []string{}
	parsingConfig := false

	for _, line := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
		if strings.Contains(line, "---") {
			// Parse config
			parsingConfig = !parsingConfig
			continue
		}
		if parsingConfig {
			pageConfig = append(pageConfig, line)
		} else {
			pageContent = append(pageContent, line)
		}
	}

	// Parse page config
	pageConfigDecoded := map[string]interface{}{}

	_, err := toml.Decode(strings.Join(pageConfig, "\n"), &pageConfigDecoded)
	if err != nil {
		log.Warnln("Couldn't parse the page config: ", err)
	}

	pageConfigDecoded["content"] = template.HTML(strings.Join(pageContent, "\n"))

	return pageConfigDecoded
}

func (k *Koekr) generatePage(file string) {

	// t *template.Template, variables map[string]interface{}

	// Read content from page
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Warnln("Couldn't read page:", file)
		return
	}

	local_variables := k.variables
	local_variables["page"] = k.parsePage(string(content))

	// Create file for output
	outputFile, err := os.Create("./generated/" + filepath.Base(file))
	if err != nil {
		log.Warnln("Couldn't create output file: ", err)
		return
	}

	// Execut template
	if err := k.t.Execute(outputFile, local_variables); err != nil {
		log.Warnln("There was an error while building the html output", err)
	}

	outputFile.Close()
}

func (k *Koekr) GenerateAllPages() {
	for _, file := range k.findPages() {
		k.generatePage(file)
	}
}

func (k *Koekr) ParseTemplates() error {
	var err error
	k.t, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatalln("Couldn't parse template files:", err)
	}
	return err
}

func (k *Koekr) ParseConfig() error {
	if _, err := toml.DecodeFile("config.toml", &k.variables); err != nil {
		log.Fatalln("Couldn't process config file:", err)
		return err
	}
	return nil
}

func main() {

	k := Koekr{}

	// Parse template files

	k.ParseTemplates()

	k.ParseConfig()

	// Create directories (if not exist)
	_ = os.Mkdir("./generated", 0755)

	// Check if 'pages' directory exist
	if _, err := os.Stat("./pages"); os.IsNotExist(err) {
		log.Fatalln("The 'pages' directory doesn't exist. So no files will be generated.")
	}

	/// Copy assets
	os.RemoveAll("./generated/assets")
	CopyDir("./assets", "./generated/assets")

	k.GenerateAllPages()

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
				if strings.Contains(event.Name, "config.toml") {
					// Parse config again and regenerated all pages
					err = k.ParseConfig()
					if err == nil {
						k.GenerateAllPages()
					}
				}
				if strings.Contains(event.Name, "index.html") {
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
	}

	// Hang so program doesn't exit
	<-done

	watcher.Close()

}
