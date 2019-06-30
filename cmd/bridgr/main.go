package main

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

// this interface may not be useful because each type of worker needs to be instantiated in main anyways
// keeping for now in case it becomes useful to abstract handling of workers
type worker interface {
	Run(config.BridgrConf) error
	Setup(config.BridgrConf) error
	Name() string
}

var verbosePtr = flag.Bool("verbose", false, "Verbose logging (debug)")
var configPtr = flag.String("config", "bridge.yml", "The config file for Bridgr (default is bridge.yml)")
var dryrunPtr = flag.Bool("dry-run", false, "Dry-run only. Do not actually download content")

func init() {
	flag.StringVar(configPtr, "c", "bridge.yml", "The config file for Bridgr (default is bridge.yml)")
	flag.BoolVar(verbosePtr, "v", false, "Verbose logging (debug)")
	flag.BoolVar(dryrunPtr, "n", false, "Dry-run only. Do not actually download content")
}

func main() {
	fmt.Println("Bridgr")
	flag.Parse()

	configFile, err := openConfig()
	if err != nil {
		log.Printf("Unable to open bridgr config \"%s\": %s", *configPtr, err)
		if configFile != nil {
			configFile.Close()
		}
		os.Exit(4)
	}
	conf, err := config.New(configFile)
	if err != nil {
		panic(err)
	}
	// spew.Dump(config)
	files := workers.NewFiles(conf)
	yum := workers.NewYum(conf)

	if *dryrunPtr {
		files.Setup()
		yum.Setup()
	} else {
		err := files.Run()
		if err != nil {
			log.Printf("Error processing files: %s", err)
		}
		err = yum.Run()
		if err != nil {
			log.Printf("Error processing Yum: %s", err)
		}
	}
}

func openConfig() (io.ReadCloser, error) {
	if !fileExists(*configPtr) {
		return nil, fmt.Errorf("file does not exist")
	}
	return os.Open(*configPtr)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
