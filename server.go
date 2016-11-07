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
func InitServers(room *Room, router *Router, handler *Handler) {
	ConsoleServ = &ConsoleServer{room: room, router: router, handler: handler}
	WifiServ = &WifiServer{room: room, router: router, handler: handler}
}
