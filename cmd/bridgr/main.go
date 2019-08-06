package main

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	verbosePtr = flag.Bool("verbose", false, "Verbose logging (debug)")
	configPtr  = flag.String("config", "bridge.yml", "The config file for Bridgr (default is bridge.yml)")
	dryrunPtr  = flag.Bool("dry-run", false, "Dry-run only. Do not actually download content")
)

func init() {
	flag.StringVar(configPtr, "c", "bridge.yml", "The config file for Bridgr (default is bridge.yml)")
	flag.BoolVar(verbosePtr, "v", false, "Verbose logging (debug)")
	flag.BoolVar(dryrunPtr, "n", false, "Dry-run only. Do not actually download content")
}

func main() {
	flag.Parse()
	bridgr.Verbose = *verbosePtr

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

	workerList := initWorkers(conf)

	subcmd := []string{"all"}
	if len(os.Args) >= 2 {
		subcmd = flag.Args()
	}
	switch subcmd[0] {
	case "docker":
		w := FindWorker(workerList, func(w workers.Worker) bool {
			return w.Name() == "Docker"
		})
		doWorker(w)
	case "yum":
		w := FindWorker(workerList, func(w workers.Worker) bool {
			return w.Name() == "Yum"
		})
		doWorker(w)
	case "files":
		w := FindWorker(workerList, func(w workers.Worker) bool {
			return w.Name() == "Files"
		})
		doWorker(w)
	case "all":
		for _, w := range workerList {
			doWorker(w)
		}
	default:
		log.Printf("Unknown subcommand `%s`", subcmd)
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

// FindWorker looks through an array of Workers to find one specified by the function
func FindWorker(items []workers.Worker, f func(workers.Worker) bool) workers.Worker {
	for _, i := range items {
		if f(i) {
			return i
		}
	}
	return nil
}

func initWorkers(conf *config.BridgrConf) []workers.Worker {
	var w []workers.Worker
	w = append(w, workers.NewYum(conf))
	w = append(w, workers.NewFiles(conf))
	w = append(w, workers.NewDocker(conf))
	return w
}

func doWorker(w workers.Worker) {
	log.Printf("Processing %s...", w.Name())
	_ = w.Setup()
	if !*dryrunPtr {
		err := w.Run()
		if err != nil {
			log.Printf("Error processing Yum: %s", err)
		}
	}
}
