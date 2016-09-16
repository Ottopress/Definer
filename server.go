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
	GetDefiner() *Definer
	GetHandler() *Handler
}

// InitServers setups up each of the servers and sets up
// their basic handlers.
func InitServers(definer *Definer) {
	ConsoleServ = NewConsoleServer(definer)
	WifiServ = NewWifiServer(definer)
}
