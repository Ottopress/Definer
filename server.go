package main

import (
	"bufio"
	"os"
)

var (
	// ConsoleServ server instance. Used for console logging/commands.
	ConsoleServ *ConsoleServer
)

// Server represents a communication system of the definer
type Server interface {
	Listen() error
	AddHandler(identifier string, handler func(server Server, core string, args ...string))
	GetDefiner() *Definer
}

// ConsoleServer repersents a console-based communication system
type ConsoleServer struct {
	Handlers map[string]func(server Server, core string, args ...string)
	Definer  *Definer
}

// InitServers setups up each of the servers and sets up
// their basic handlers.
func InitServers(definer *Definer) {
	ConsoleServ = NewConsoleServer(definer)
	setupHandlers()
}

// setupHandlers adds all the pre-established handler functions
// to their respective servers.
// NOTE: THIS IS NOT AN AUTOMATIC PROCESS. THIS IS DONE MANUALLY.
func setupHandlers() {
	ConsoleServ.AddHandler("router", HandleRouter)
}

// NewConsoleServer returns a new ConsoleServer
func NewConsoleServer(definer *Definer) *ConsoleServer {
	return &ConsoleServer{
		Handlers: make(map[string]func(server Server, core string, args ...string)),
		Definer:  definer,
	}
}

// Listen begins listening for console commands that have been registered
// in the handlers.
func (console *ConsoleServer) Listen() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmdArgs := ToArgv(scanner.Text())
		console.Handlers[cmdArgs[0]](console, cmdArgs[0], cmdArgs[1:]...)
	}
	scannerErr := scanner.Err()
	if scannerErr != nil {
		return scannerErr
	}
	return nil
}

// AddHandler links a command to a handler function
func (console *ConsoleServer) AddHandler(identifier string, handler func(server Server, core string, args ...string)) {
	console.Handlers[identifier] = handler
}

// GetDefiner returns the provided Definer instance
func (console *ConsoleServer) GetDefiner() *Definer {
	return console.Definer
}
