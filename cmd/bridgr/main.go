package main

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"flag"
	"fmt"
)

// this interface may not be useful because each type of worker needs to be instantiated in main anyways
// keeping for now in case it becomes useful to abstract handling of workers
type worker interface {
	Run(config.BridgrConf) error
	Setup(config.BridgrConf) error
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

	configFile, err := config.Config(*configPtr)
	if err != nil {
		panic(err)
	}
	// spew.Dump(config)

	files := workers.Files{}
	if *dryrunPtr {
		files.Setup(configFile)
	} else {
		files.Run(configFile)
	}
}
