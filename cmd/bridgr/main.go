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

	if *dryrunPtr {
		log.Println("Dry-Run requested, will not download artifacts.")
	}

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

	subcmd := flag.Args()
	if len(subcmd) <= 0 {
		subcmd = []string{"all"}
	}
	bridgr.Debugf("Running workers for subcommands: %+v\n", subcmd)
	err = processWorkers(workerList, subcmd[0])
	if err != nil {
		log.Print(err)
		os.Exit(255)
	}
	os.Exit(0)
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
	return []workers.Worker{
		workers.NewYum(conf),
		workers.NewFiles(conf),
		workers.NewDocker(conf),
	}
}

func doWorker(w workers.Worker) {
	log.Printf("Processing %s...", w.Name())
	var err error
	if *dryrunPtr {
		err = w.Setup()
	} else {
		err = w.Run()
	}
	if err != nil {
		log.Printf("Error processing %s: %s", w.Name(), err)
	}
}

func processWorkers(list []workers.Worker, filter string) error {
	// TODO: This only works on a single subcommand right now. Allow this to work on an array of subcommands.
	switch filter {
	case "docker":
		w := FindWorker(list, func(w workers.Worker) bool {
			return w.Name() == "Docker"
		})
		doWorker(w)
	case "yum":
		w := FindWorker(list, func(w workers.Worker) bool {
			return w.Name() == "Yum"
		})
		doWorker(w)
	case "files":
		w := FindWorker(list, func(w workers.Worker) bool {
			return w.Name() == "Files"
		})
		doWorker(w)
	case "all":
		for _, w := range list {
			doWorker(w)
		}
	default:
		log.Printf("Unknown subcommand `%s`", filter)
	}
	return nil
}
