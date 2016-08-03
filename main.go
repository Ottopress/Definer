package main

import (
	"os"
)

const (
	// EnvPhysical indicates that the environment is the
	// physical device
	EnvPhysical int = iota
	// EnvEmulated indicates that the environemtn is emulated,
	// probably for development
	EnvEmulated
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
    Info.Println("Ottopress Definer starting...")
    Info.Println("Initializing Definer...")
    definer, definerErr := InitDefiner(DefinerPath)
    if definerErr != nil {
        Error.Println(definerErr.Error())
        os.Exit(1)
    }

	routerInitErr := definer.Router.Initialize()
	if routerInitErr != nil {
		Error.Println(routerInitErr.Error())
		os.Exit(1)
	}
	Info.Println("Router interface successfully initialized.")
	Debug.Println(definer.Router.Interface.Name)
	Info.Println("Preparing to connect to \"" + definer.Router.SSID + "\"...")
	routerConnErr := definer.Router.Connect()
	if routerConnErr != nil {
		Error.Println(routerConnErr.Error())
		os.Exit(1)
	}
	Info.Println("Connection successful!")
	Info.Println("Waiting...")
	console := NewConsoleServer(definer)
	console.AddHandler("config-router", HandleConfigRouter)
	go console.Listen()
	for {}
}

// HandleConfigRouter lol
func HandleConfigRouter(server Server, core string, args ...string) {
	shouldSkip := false
	for index, item := range args {
		if shouldSkip {
			shouldSkip = false
			continue
		}
		switch item {
		case "ssid":
			server.GetDefiner().Router.SSID = args[index+1]
			shouldSkip = true
		case "password":
			server.GetDefiner().Router.Password = args[index+1]
			shouldSkip = true
		}
	}
	Debug.Println(server.GetDefiner().Router)
}
