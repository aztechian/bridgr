package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	log "unknwon.dev/clog/v2"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/aztechian/bridgr/internal/bridgr/cmd"
)

const (
	success = 0
	execErr = 1
	cfgErr  = 4
	srvErr  = 255

	defaultTimeout = time.Second * 20
)

var (
	verbosePtr     = flag.Bool("verbose", false, "Verbose logging (debug)")
	versionPtr     = flag.Bool("version", false, "Print version and exit")
	hostPtr        = flag.Bool("host", false, "Run Bridgr in hosting mode. This only runs a web server for \"packages\" directory")
	hostListenPtr  = flag.String("listen", ":8080", "Listen address for Bridger. Only applicable in hosting mode.")
	configPtr      = flag.String("config", "bridge.yaml", "The config file for Bridgr (default is bridge.yaml)")
	threadsPtr     = flag.Int("threads", 1, "Number of threads to use for fetching artifacts")
	dryrunPtr      = flag.Bool("dry-run", false, "Dry-run only. Do not actually download content")
	fileTimeoutPtr = flag.Duration("file-timeout", defaultTimeout, "Timeout duration for downloading files, uses Golang duration strings")
)

func init() {
	flag.StringVar(configPtr, "c", "bridge.yaml", "The config file for Bridgr (default is bridge.yaml)")
	flag.BoolVar(verbosePtr, "v", false, "Verbose logging (debug)")
	flag.BoolVar(hostPtr, "H", false, "Run Bridgr in hosting mode. This only runs a web server for \"packages\" directory")
	flag.StringVar(hostListenPtr, "l", ":8080", "Listen address for Bridger. Only applicable in hosting mode.")
	flag.IntVar(threadsPtr, "t", runtime.NumCPU(), "Number of threads to use for fetching artifacts")
	flag.BoolVar(dryrunPtr, "n", false, "Dry-run only. Do not actually download content")
	flag.DurationVar(fileTimeoutPtr, "x", defaultTimeout, "Timeout duration for downloading files, uses Golang duration strings")
}

func main() {
	flag.Parse()
	logConfig := log.ConsoleConfig{Level: log.LevelWarn}
	if *verbosePtr {
		logConfig.Level = log.LevelTrace
	}
	// initialize clog logger
	if err := log.NewConsole(1, logConfig); err != nil {
		panic("unable to create console logger: " + err.Error())
	}

	if *versionPtr {
		fmt.Fprintln(os.Stderr, "Bridgr - (C) 2020 Ian Martin, MIT License. See https://github.com/aztechian/bridgr")
		fmt.Printf("%s\n", bridgr.Version)
		fmt.Fprintln(os.Stderr, "")
		exit(success)
	}

	if *hostPtr {
		dir := http.Dir(bridgr.BaseDir(""))
		err := bridgr.Serve(*hostListenPtr, dir)
		if err != nil {
			log.Error("Unable to start HTTP Server: %s\n", err)
			exit(srvErr)
		}
		exit(success)
	}

	if *dryrunPtr {
		bridgr.DryRun = *dryrunPtr
		log.Info("Dry-Run requested, will not download artifacts.")
	}

	if fileTimeoutPtr != nil {
		bridgr.FileTimeout = *fileTimeoutPtr
		log.Trace("setting file timeout to %s", *fileTimeoutPtr)
	}

	configFile, err := openConfig()
	if err != nil {
		log.Error("Unable to open bridgr config \"%s\": %s", *configPtr, err)
		if configFile != nil {
			configFile.Close()
		}
		exit(cfgErr)
	}
	config, err := cmd.New(configFile)
	if err != nil {
		log.Info(err.Error())
		exit(execErr)
	}

	if err := config.Execute(flag.Args()); err != nil {
		log.Info(err.Error())
		exit(execErr)
	}
	exit(success)
}

func exit(code int) {
	log.Stop()
	os.Exit(code)
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
