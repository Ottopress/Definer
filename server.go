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
}

// InitServers setups up each of the servers and sets up
// their basic handlers.
func InitServers(router *Router, handler *Handler, deviceManager *DeviceManager) {
	ConsoleServ = &ConsoleServer{handler: handler, router: router, deviceManager: deviceManager}
	WifiServ = &WifiServer{handler: handler, router: router}
}
