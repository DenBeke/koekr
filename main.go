package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

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
