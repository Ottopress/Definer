package main

import (
	"bufio"
	"os"
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

// NewConsoleServer returns a new ConsoleServer
func NewConsoleServer(Definer *Definer) *ConsoleServer {
	return &ConsoleServer{
		Handlers: make(map[string]func(server Server, core string, args ...string)),
		Definer:  Definer,
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
