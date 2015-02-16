package main

import (
	"github.com/aidanhs/go-dockerclient"
	flag "github.com/docker/docker/pkg/mflag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
		fmt.Fprintf(helpFd, "  The DOCKER_HOST environment variable overrides the default location to find the docker daemon\n")
	}

	flag.BoolVar(&doHelp, []string{"h", "-help"}, false, "Pring this help text")
	flag.StringVar(&listenOn, []string{"l"}, "127.0.0.1:5000", "Address to listen on")
	flag.StringVar(&outDir, []string{"o"}, ".", "Directory to store data in")
	flag.BoolVar(&doDebug, []string{"-debug"}, false, "Set log level to debug")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	imgId := flag.Arg(0)
	if len(imgId) == 0 {
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
	imgJson, err := client.InspectImage(imgId)
	if err != nil {
		logger.Error("%s", err.Error())
		os.Exit(1)
	}
	layerId = imgJson.ID
	if layerId != imgId {
		logger.Info("Full layer id found: %s", layerId)
	}
	layerLock.Unlock()

	logger.Debug("Starting shim registry on %s", listenOn)
	go (func () {
		if err := http.ListenAndServe(listenOn, NewHandler(outDir)); err != nil {
			logger.Error("%s", err.Error())
			os.Exit(1)
		}
	})()

	sleeps := []int{1, 5, 10, 100, 200, 500, 1000, 2000}
	pingUrl := "http://" + listenOn + "/v1/_ping"
	apiIsUp := false
	logger.Debug("Waiting for shim registry to start by checking %s", pingUrl)
	for _, ms := range sleeps {
		logger.Debug("Sleeping %d ms before ping", ms)
		time.Sleep(time.Duration(ms) * time.Millisecond)
		resp, err := http.Get(pingUrl)
		if err == nil {
			resp.Body.Close()
			apiIsUp = true
			break
		}
	}
	if !apiIsUp {
		logger.Error("Shim registry took too long to come up")
		os.Exit(1)
	}

	logger.Debug("Shim Registry Started")

	err = dockerMain(client, listenOn)
	if err != nil {
		logger.Error("%s", err)
		os.Exit(1)
	}
}

func dockerMain(client *docker.Client, regUrl string) (err error) {
	imgName := regUrl + "/" + "dlgrab_push_staging_tmp"
	imgTag := "latest"

	logger.Debug("Tagging image into temporary repo")
	tagOpts := docker.TagImageOptions{
		Repo: imgName,
		Tag: imgTag,
		Force: false,
	}
	layerLock.Lock()
	err = client.TagImage(layerId, tagOpts)
	layerLock.Unlock()
	if err != nil {
		return
	}

	logger.Debug("Pushing image")
	pushOpts := docker.PushImageOptions{
		Registry: "",
		Name: imgName,
		Tag: imgTag,
	}
	err = client.PushImage(pushOpts, docker.AuthConfiguration{})
	if err != nil {
		return
	}

	logger.Debug("Removing temporary image tag")
	removeOpts := docker.RemoveImageOptions{
		Force: false,
		NoPrune: true,
	}
	err = client.RemoveImage(imgName, removeOpts)
	if err != nil {
		return
	}
	return nil
}
