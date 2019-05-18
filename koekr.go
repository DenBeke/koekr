package main

import (
	"bytes"
	template "text/template"
	xmlTemplate "text/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type Koekr struct {
	variables map[string]interface{}
	t         *template.Template
	files     []string

	config struct {
		watch      bool
		configFile string
		template   string
	}
}

func (k *Koekr) findPages() []string {
	searchDir := "./pages/"

	_ = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {

		fi, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if fi.Mode().IsRegular() {
			k.files = append(k.files, path)
		}

		return nil
	})

	return k.files
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

	pageConfigDecoded["content"] = strings.Join(pageContent, "\n")

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

	pageVariables := k.parsePage(string(content))

	local_variables := k.variables
	local_variables["page"] = pageVariables


	// Process sub template. This allows page templates to be actual Go templates
	executedContent := bytes.Buffer{}
	localTemplate := template.New(file)
	templateString, _ := pageVariables["content"].(string)
	localTemplate.Parse(string(templateString))
	err = localTemplate.Execute(&executedContent, local_variables)
	if err != nil {
		log.Warnln("There was an error while building the html output for a page", err)
	}
	pageVariables["content"] = executedContent.String()
	local_variables["page"] = pageVariables



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
	k.findPages()
	for _, file := range k.files {
		k.generatePage(file)
	}
	k.GenerateSitemap()
}

func (k *Koekr) GenerateSitemap() {
	
	sitemapTemplate := 
`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	{{ $url := .Variables.site.url }}
	{{ range $slug := .Slugs }}
   <url>
	  <loc>{{ $url }}/{{ $slug }}</loc>
   </url>
   {{ end }}
</urlset>
`
	
	outputFile, err := os.Create("./generated/sitemap.xml")
	if err != nil {
		log.Warnln("Couldn't create sitemap.xml file: ", err)
		return
	}
	
	t, err := xmlTemplate.New("sitemap").Parse(sitemapTemplate)
	if err != nil {
		log.Warnln("Couldn't generate sitemap:", err)
		return
	}
	
	slugs := []string{}
	
	for _, file := range k.files {
		slugs = append(slugs, filepath.Base(file))
	}
	data := struct{
		Slugs []string
		Variables map[string]interface{} 
	}{
		Slugs: slugs,
		Variables: k.variables,
	}
	if err := t.Execute(outputFile, data); err != nil {
		log.Warnln("There was an error while building the sitemap output", err)
	}
}

func (k *Koekr) ParseTemplates() error {
	var err error
	k.t, err = template.New("index.html").Funcs(template.FuncMap(sprig.FuncMap())).ParseFiles(k.config.template)
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
