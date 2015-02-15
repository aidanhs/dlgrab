package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var logger = &Logger{}

var (
	GITCOMMIT string
)

func startServer(listenOn, outDir string) {
	logger.Info("using version ", GITCOMMIT)
	logger.Info("starting server on ", listenOn)
	logger.Info("using outDir ", outDir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
	go func() {
		sig := <-c
		logger.Debug("Received signal '%v', exiting\n", sig)
		os.Exit(0)
	}()

	if err := http.ListenAndServe(listenOn, NewHandler(outDir)); err != nil {
		logger.Error(err.Error())
	}
}

func main() {
	var listenOn *string
	var outDir *string
	var doDebug *bool

	listenOn = flag.String("l", ":5000", "Address on which to listen.")
	outDir = flag.String("d", ".", "Directory to store data in")
	doDebug = flag.Bool("D", false, "set log level to debug")
	flag.Parse()

	logger.Level = INFO
	if *doDebug {
		logger.Level = DEBUG
	}
	startServer(*listenOn, *outDir)
}
