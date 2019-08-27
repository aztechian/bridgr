package main

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	version    = "development"
	verbosePtr = flag.Bool("verbose", false, "Verbose logging (debug)")
	versionPtr = flag.Bool("version", false, "Print version and exit")
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

	if *versionPtr {
		fmt.Fprintln(os.Stderr, "Bridgr - (C) 2019 Ian Martin, MIT License. See https://github.com/aztechian/bridgr")
		fmt.Printf("%s\n", version)
		fmt.Fprintln(os.Stderr, "")
		os.Exit(0)
	}

	if *dryrunPtr {
		bridgr.Println("Dry-Run requested, will not download artifacts.")
	}

	configFile, err := openConfig()
	if err != nil {
		bridgr.Printf("Unable to open bridgr config \"%s\": %s", *configPtr, err)
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

	bridgr.Debugf("Running workers for subcommands: %+v\n", flag.Args())
	err = processWorkers(workerList, flag.Args())
	if err != nil {
		bridgr.Print(err)
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
		workers.NewPython(conf),
	}
}

func doWorker(w workers.Worker) {
	bridgr.Printf("Processing %s...", w.Name())
	var err error
	if *dryrunPtr {
		err = w.Setup()
	} else {
		err = w.Run()
	}
	if err != nil {
		bridgr.Printf("Error processing %s: %s", w.Name(), err)
	}
}

func processWorkers(list []workers.Worker, filter []string) error {
	// TODO: This only works on a single subcommand right now. Allow this to work on an array of subcommands.
	if len(filter) <= 0 {
		filter = append(filter, "all")
	}
	switch f := filter[0]; f {
	case "all", "":
		for _, w := range list {
			doWorker(w)
		}
	default:
		w := FindWorker(list, func(w workers.Worker) bool {
			return strings.EqualFold(w.Name(), f)
		})
		if w == nil {
			return fmt.Errorf("Unable to find worker named %s", f)
		}
		doWorker(w)
	}
	return nil
}
