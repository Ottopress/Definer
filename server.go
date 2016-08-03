package main

import (
	"bufio"
	"os"
)

type Server interface {
	Listen() error
	AddHandler(identifier string, handler func(server Server, core string, args ...string))
	GetDefiner() *Definer
}

type ConsoleServer struct {
	Handlers map[string]func(server Server, core string, args ...string)
	Definer  *Definer
}

func NewConsoleServer(Definer *Definer) *ConsoleServer {
	return &ConsoleServer{
		Handlers: make(map[string]func(server Server, core string, args ...string)),
		Definer:  Definer,
	}
}

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

func (console *ConsoleServer) AddHandler(identifier string, handler func(server Server, core string, args ...string)) {
	console.Handlers[identifier] = handler
}

func (console *ConsoleServer) GetDefiner() *Definer {
	return console.Definer
}
