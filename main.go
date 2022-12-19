package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	commandLine    = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	logLevelFlag   = commandLine.String("log-level", "debug", "Logging level [quiet|debug|info|warning|error]")
	kubeConfigFile = commandLine.String("kubeconfig", "", "If set, the rotator used the kubeconfig file instead of in-cluster configuration.")
	varnishWorkDir = commandLine.String("n", "/etc/varnish/work/", "Varnish working folder")
	varnishVclFile = commandLine.String("f", "/etc/varnish/default.vcl", "Varnish default vcl")
)

func initLogger() {
	InitNewLogger(os.Stdout, GetLogLevelID(*logLevelFlag))
}

func parseCliArgs() error {
	commandLine.Usage = func() {
		fmt.Println("varnish-controller usage:")
		fmt.Println()
		commandLine.PrintDefaults()
	}

	if err := commandLine.Parse(os.Args[1:]); err != nil {
		return err
	}

	return nil
}

func main() {
	err := parseCliArgs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	initLogger()
	initBackendStore(*varnishWorkDir, *varnishVclFile)

	k8sSession, err := initK8s(*kubeConfigFile)
	if err != nil {
		Error(err)
		os.Exit(1)
	}

	k8sSession.startInformers()
	select {}
}
