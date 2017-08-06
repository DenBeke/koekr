package main

import (
    "html/template"
    "os"
    "github.com/BurntSushi/toml"
    "path/filepath"
    "io/ioutil"
    "strings"
    log "github.com/sirupsen/logrus"
)


func findPages() (files []string) {
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


func parsePage(content string)  map[string]interface{} {
    // Split page config from page content
    pageConfig := []string{}
    pageContent := []string{}
    parsingConfig := false
    
    for _, line := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
        if(strings.Contains(line, "---")) {
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


func generatePage(file string, t *template.Template, variables map[string]interface{}) {
    
    // Read content from page
    content, err := ioutil.ReadFile(file)
    if err != nil {
        log.Warnln("Couldn't read page:", file)
        return
    }
    
    
    variables["page"] = parsePage(string(content))
    
    
    // Create file for output
    outputFile, err := os.Create("./generated/" + filepath.Base(file))
    if err != nil {
        log.Warnln("Couldn't create output file: ", err)
        return
    }
    
    // Execut template
    if err := t.Execute(outputFile, variables); err != nil {
        log.Warnln("There was an error while building the html output", err)
    }
    
    outputFile.Close()
}


func main() {
    
    // Parse template files
    t, err := template.ParseFiles("index.html")
    if err != nil {
        log.Fatalln("Couldn't parse template files:", err)
    }

    variables := map[string]interface{}{}
    if _,err := toml.DecodeFile("config.toml", &variables); err != nil {
        log.Fatalln("Couldn't process config file:", err)
    }
    
    
    // Create directories
    _ = os.Mkdir("./generated", 0755)
    
    // Check if 'pages' directory exist
    if _, err := os.Stat("./pages"); os.IsNotExist(err) {
        log.Fatalln("The 'pages' directory doesn't exist. So no files will be generated.")
    }
    
    
    for _,file := range findPages() {
        
        generatePage(file, t, variables)
        
    }

    
}