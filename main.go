package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
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

	// configPath is the path of the config file
	configPath = "./config.xml"
	// config is the initialized config struct
	config *Config
	// room is the room of the current Definer
	room *Room
	// router is the router of the current Definer
	router *Router
	// deviceManager contains the devices the
	// current definer controls
	deviceManager *DeviceManager
	// routerManager contains the routers the
	// current definer can access
	routerManager *RouterManager
	// Environment represents the environment this software
	// is running under
	Environment = EnvEmulated
)

func main() {
	start := time.Now()
	InitLog(debugOut, infoOut, warningOut, errorOut)
	Info.Println(OttopressHeader)
	Info.Println("Definer starting...")
	Info.Println("Loading Config...")
	config, configErr := InitConfig(configPath)
	if configErr != nil {
		Error.Println(configErr.Error())
		os.Exit(1)
	}
	Info.Println("Config loaded!")
	router = config.Router
	room = config.Room
	deviceManager = config.DeviceManager
	routerManager = config.RouterManager
	Info.Println("Initializing Cleanup Handler...")
	InitCleanup(config)
	Info.Println("Cleanup Handler initialized!")
	Info.Println("Initialize Servers...")
	handler := &Handler{room, router, deviceManager, routerManager}
	InitServers(room, router, handler, deviceManager)
	go ConsoleServ.Listen()
	go WifiServ.Listen()
	Info.Println("Servers initialized!")
	Info.Println("Initializing Router...")
	routerInitErr := config.Router.Initialize()
	if routerInitErr != nil {
		Error.Println(routerInitErr)
	}
	Info.Println("Router initialized!")
	elapsed := time.Since(start).Seconds()
	Info.Printf("Done! [took %.3f seconds]...", elapsed)
	for {
	}
}

// InitCleanup initializes the cleanup handler
func InitCleanup(config *Config) {
	chanInterrupt := make(chan os.Signal, 1)
	signal.Notify(chanInterrupt, os.Interrupt)
	signal.Notify(chanInterrupt, syscall.SIGTERM)
	go func() {
		<-chanInterrupt
		cleanup(config)
		os.Exit(1)
	}()
}

// cleanup ensures that all connections are closed,
// files are written, etc. before the software restarts
func cleanup(config *Config) {
	writeErr := config.WriteConfig(configPath)
	if writeErr != nil {
		Error.Println(writeErr.Error())
		return
	}
}
