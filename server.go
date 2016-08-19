package main

var (
	// ConsoleServ server instance. Used for console logging/commands.
	ConsoleServ *ConsoleServer
	// WifiServ server instance. Used for wifi communication.
	WifiServ *WifiServer
)

// Server represents a communication system of the definer
type Server interface {
	Listen()
	AddHandler(identifier string, handler func(server Server, core string, args ...string))
	GetDefiner() *Definer
}

// InitServers setups up each of the servers and sets up
// their basic handlers.
func InitServers(definer *Definer) {
	ConsoleServ = NewConsoleServer(definer)
	WifiServ = NewWifiServer(definer)
	setupHandlers()
}

// setupHandlers adds all the pre-established handler functions
// to their respective servers.
// NOTE: THIS IS NOT AN AUTOMATIC PROCESS. THIS IS DONE MANUALLY.
func setupHandlers() {
	ConsoleServ.AddHandler("router", HandleRouter)
}
