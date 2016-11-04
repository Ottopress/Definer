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
	GetRoom() *Room
	GetRouter() *Router
	GetHandler() *Handler
}

// InitServers setups up each of the servers and sets up
// their basic handlers.
func InitServers(room *Room, router *Router, handler *Handler) {
	ConsoleServ = NewConsoleServer(room, router, handler)
	WifiServ = NewWifiServer(room, router, handler)
}
