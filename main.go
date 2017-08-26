package main

import (
	"flag"
	"github.com/BurntSushi/toml"
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

	config struct {
		watch      bool
		configFile string
		template   string
	}
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
	k.t, err = template.ParseFiles(k.config.template)
	if err != nil {
		log.Fatalln("Couldn't parse template files:", err)
	}
	return err
}

func (k *Koekr) ParseConfig() error {
	if _, err := toml.DecodeFile(k.config.configFile, &k.variables); err != nil {
		log.Fatalln("Couldn't process config file:", err)
		return err
	}
	return nil
}

func main() {

	k := Koekr{}

	// Command line flags
	flag.BoolVar(&k.config.watch, "watch", false, "Watch for file changes?")
	flag.StringVar(&k.config.configFile, "config", "config.toml", "Config file")
	flag.StringVar(&k.config.template, "template", "index.html", "Template files")
	flag.Parse()

	// Parse template files
	k.ParseTemplates()

	// Parse config
	k.ParseConfig()

	// Create directories (if not exist)
	_ = os.Mkdir("./generated", 0755)

	// Check if 'pages' directory exist
	if _, err := os.Stat("./pages"); os.IsNotExist(err) {
		log.Fatalln("The 'pages' directory doesn't exist. So no files will be generated.")
	}

	// Generate the actual pages
	k.GenerateAllPages()

	// Copy assets
	os.RemoveAll("./generated/assets")
	CopyDir("./assets", "./generated/assets")

	// Watch for file changes
	if k.config.watch {
		k.WatchForChanges()
	}

}
