package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/aztechian/bridgr/internal/bridgr/cmd"
)

var (
	verbosePtr    = flag.Bool("verbose", false, "Verbose logging (debug)")
	versionPtr    = flag.Bool("version", false, "Print version and exit")
	hostPtr       = flag.Bool("host", false, "Run Bridgr in hosting mode. This only runs a web server for \"packages\" directory")
	hostListenPtr = flag.String("listen", ":8080", "Listen address for Bridger. Only applicable in hosting mode.")
	configPtr     = flag.String("config", "bridge.yml", "The config file for Bridgr (default is bridge.yml)")
	threadsPtr    = flag.Int("threads", 1, "Number of threads to use for fetching artifacts")
	dryrunPtr     = flag.Bool("dry-run", false, "Dry-run only. Do not actually download content")
)

func init() {
	flag.StringVar(configPtr, "c", "bridge.yml", "The config file for Bridgr (default is bridge.yml)")
	flag.BoolVar(verbosePtr, "v", false, "Verbose logging (debug)")
	flag.BoolVar(hostPtr, "H", false, "Run Bridgr in hosting mode. This only runs a web server for \"packages\" directory")
	flag.StringVar(hostListenPtr, "l", ":8080", "Listen address for Bridger. Only applicable in hosting mode.")
	flag.IntVar(threadsPtr, "t", runtime.NumCPU(), "Number of threads to use for fetching artifacts")
	flag.BoolVar(dryrunPtr, "n", false, "Dry-run only. Do not actually download content")
}

func main() {
	flag.Parse()
	bridgr.Verbose = *verbosePtr

	if *versionPtr {
		fmt.Fprintln(os.Stderr, "Bridgr - (C) 2020 Ian Martin, MIT License. See https://github.com/aztechian/bridgr")
		fmt.Printf("%s\n", bridgr.Version)
		fmt.Fprintln(os.Stderr, "")
		os.Exit(0)
	}

	if *hostPtr {
		dir := http.Dir(bridgr.BaseDir(""))
		err := bridgr.Serve(*hostListenPtr, dir)
		if err != nil {
			fmt.Printf("Unable to start HTTP Server: %s\n", err)
			os.Exit(255)
		}
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
	config, err := cmd.New(configFile)
	if err != nil {
		panic(err)
	}
	cmd.Execute(*config, flag.Args())
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
