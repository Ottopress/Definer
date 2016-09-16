package main

import (
	"os"
	"os/signal"
	"syscall"
)

const (
	// EnvPhysical indicates that the environment is the
	// physical device
	EnvPhysical int = iota
	// EnvEmulated indicates that the environemtn is emulated,
	// probably for development
	EnvEmulated
)

const (
	// OttopressHeader ...
	OttopressHeader = `
    _____  _______ _______  _____   _____   ______ _______ _______ _______
   |     |    |       |    |     | |_____] |_____/ |______ |______ |______
   |_____|    |       |    |_____| |       |    \_ |______ ______| ______|
=============================================================================`
)

var (
	// debugOut redirects debug logs to os.Stdout
	debugOut = os.Stdout
	// infoOut redirects info logs to os.Stdout
	infoOut = os.Stdout
	// warningOut redirects warning logs to os.Stdout
	warningOut = os.Stdout
	// errorOut redirects error logs to os.Stderr
	errorOut = os.Stderr

	// DefinerPath is the path of the config for the
	// definer settings.
	DefinerPath = "./definer.xml"
	// SelfDefiner is the definer struct representing the
	// definer.
	SelfDefiner Definer
	// Environment represents the environment this software
	// is running under
	Environment = EnvEmulated
)

func main() {
	InitLog(debugOut, infoOut, warningOut, errorOut)
	Info.Println(OttopressHeader)
	Info.Println("Ottopress Definer starting...")
	Info.Println("Initializing Definer...")
	definer, definerErr := InitDefiner(DefinerPath)
	if definerErr != nil {
		Error.Println(definerErr.Error())
		os.Exit(1)
	}
	Info.Println("Definer initialized!")
	Info.Println("Initializing Cleanup Handler...")
	InitCleanup(definer)
	Info.Println("Cleanup Handler initialized!")
	Info.Println("Initialize Servers...")
	InitServers(definer)
	go ConsoleServ.Listen()
	go WifiServ.Listen()
	Info.Println("Servers initialized!")
	Info.Println("Initializing Router...")
	routerInitErr := definer.Router.Initialize()
	if routerInitErr != nil {
		Error.Println(routerInitErr)
	}
	Info.Println("Router initialized!")
	Info.Println("Waiting...")
	for {
	}
}

// InitCleanup initializes the cleanup handler
func InitCleanup(definer *Definer) {
	chanInterrupt := make(chan os.Signal, 1)
	signal.Notify(chanInterrupt, os.Interrupt)
	signal.Notify(chanInterrupt, syscall.SIGTERM)
	go func() {
		<-chanInterrupt
		cleanup(definer)
		os.Exit(1)
	}()
}

// cleanup ensures that all connections are closed,
// files are written, etc. before the software restarts
func cleanup(definer *Definer) {
	writeErr := definer.WriteDefiner(DefinerPath)
	if writeErr != nil {
		Error.Println(writeErr.Error())
		return
	}
}
