package main

import (
	"github.com/fsouza/go-dockerclient"
	flag "github.com/docker/docker/pkg/mflag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var layerLock sync.Mutex
var layerId string = ""

var logger = &Logger{}

var (
	GITCOMMIT string
)

func main() {
	var listenOn string
	var outDir string
	var doDebug bool
	var doHelp bool

	helpFd := os.Stderr
	flag.Usage = func () {
		fmt.Fprintf(helpFd, "Usage for %s [flags...] LAYER\n", os.Args[0])
		fmt.Fprintf(helpFd, "  LAYER: layer id to export, or image name to export top layer of\n")
		flag.PrintDefaults()
		fmt.Fprintf(helpFd, "\n")
		fmt.Fprintf(helpFd, "  The DOCKER_HOST environment variable overrides the default location to find the docker daemon")
	}

	flag.BoolVar(&doHelp, []string{"h", "-help"}, false, "Pring this help text")
	flag.StringVar(&listenOn, []string{"l"}, "127.0.0.1:5000", "Address on which to listen")
	flag.StringVar(&outDir, []string{"o"}, ".", "Directory to store data in")
	flag.BoolVar(&doDebug, []string{"-debug"}, false, "Set log level to debug")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	imageId := flag.Arg(0)
	if len(imageId) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	if doHelp {
		helpFd = os.Stdout
		flag.Usage()
		return
	}
	logger.Level = INFO
	if doDebug {
		logger.Level = DEBUG
	}

	logger.Debug("DLGrab version %s", GITCOMMIT)

	logger.Info("Putting output in %s", outDir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
	go func() {
		sig := <-c
		logger.Debug("Received signal '%v', exiting\n", sig)
		os.Exit(1)
	}()

	endpoint := os.Getenv("DOCKER_HOST")
	if endpoint == "" {
		endpoint = "unix:///var/run/docker.sock"
	}
	client, err := docker.NewClient(endpoint)
	if err != nil {
		logger.Error("%s", err.Error())
		os.Exit(1)
	}

	layerLock.Lock()
	imgJson, err := client.InspectImage(imageId)
	if err != nil {
		logger.Error("%s", err.Error())
		os.Exit(1)
	}
	layerId = imgJson.ID
	if layerId != imageId {
		logger.Info("Full layer id found: %s", layerId)
	}
	layerLock.Unlock()

	logger.Debug("Starting shim registry on %s", listenOn)
	if err := http.ListenAndServe(listenOn, NewHandler(outDir)); err != nil {
		logger.Error("%s", err.Error())
		os.Exit(1)
	}
}
