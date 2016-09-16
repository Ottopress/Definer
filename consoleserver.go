package main

import (
	"bufio"
	"errors"
	"os"

	"github.com/golang/protobuf/proto"

	"git.getcoffee.io/ottopress/definer/protos"
)

// ConsoleServer repersents a console-based communication system
type ConsoleServer struct {
	Handler *Handler
	Definer *Definer
}

// ConsoleOut represents a console-based output. This is used
// instead of Stdin as an intermediary in order to produce
// human readable output.
type ConsoleOut struct{}

// NewConsoleServer returns a new ConsoleServer
func NewConsoleServer(definer *Definer) *ConsoleServer {
	console := &ConsoleServer{
		Definer: definer,
	}
	handler := &Handler{
		Server: console,
	}
	console.Handler = handler
	return console
}

// Listen begins listening for console commands that have been registered
// in the handlers.
func (console *ConsoleServer) Listen() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmdArgs := ToArgv(scanner.Text())
		packet, packetErr := console.toProto(cmdArgs...)
		if packetErr != nil {
			Error.Println("console: error parsing command: " + packetErr.Error())
			return
		}
		handleErr := console.Handler.Handle(os.Stdout, packet)
		if handleErr != nil {
			Error.Println("console: error handling command: " + handleErr.Error())
		}
	}
	scannerErr := scanner.Err()
	if scannerErr != nil {
		Error.Println("console: scanner encountered error: " + scannerErr.Error())
		return
	}
	Error.Println("console: an unexpected error occured")
	return
}

//
func (console *ConsoleServer) toProto(args ...string) (*packets.Wrapper, error) {
	if (len(args) % 2) != 1 {
		return nil, errors.New("console: invalid number of arguments")
	}
	var argMap = make(map[string]string)
	for i := 1; i < (len(args) - 1); i += 2 {
		argMap[args[i]] = args[i+1]
	}
	switch args[0] {
	case "introductionserver":
		return console.buildIntroductionServer(argMap)
	case "routerconfig":
		return console.buildRouterRequest(argMap)
	}
	return nil, nil
}

func (console *ConsoleServer) buildIntroductionServer(args map[string]string) (*packets.Wrapper, error) {
	var setup bool
	if args["setup"] == "true" {
		setup = true
	} else {
		setup = false
	}
	return &packets.Wrapper{
		Header: &packets.Wrapper_Header{
			Origin:      "127.0.0.1",
			Destination: "127.0.0.1",
			Id:          "0",
			Type:        packets.Wrapper_Header_SERVER,
		},
		Body: &packets.Wrapper_Intro{
			Intro: &packets.IntroductionServer{
				Setup: setup,
			},
		},
	}, nil
}

func (console *ConsoleServer) buildRouterRequest(args map[string]string) (*packets.Wrapper, error) {
	return &packets.Wrapper{
		Header: &packets.Wrapper_Header{
			Origin:      "127.0.0.1",
			Destination: "127.0.0.1",
			Id:          "0",
			Type:        packets.Wrapper_Header_REQUEST,
		},
		Body: &packets.Wrapper_RouterConfigReq{
			RouterConfigReq: &packets.RouterConfigurationRequest{
				Ssid:     args["ssid"],
				Password: args["password"],
			},
		},
	}, nil
}

// GetDefiner returns the provided Definer instance
func (console *ConsoleServer) GetDefiner() *Definer {
	return console.Definer
}

// GetHandler returns the provided Handler instance
func (console *ConsoleServer) GetHandler() *Handler {
	return console.Handler
}

func (consoleOut *ConsoleOut) Write(p []byte) (int, error) {
	proto, protoErr := consoleOut.parseProto(p[3 : len(p)-1])
	if protoErr != nil {
		return 0, protoErr
	}
	Info.Println(proto)
	return len(p), nil
}

func (consoleOut *ConsoleOut) parseProto(protoData []byte) (packets.Wrapper, error) {
	var wrapper packets.Wrapper
	unmarshErr := proto.Unmarshal(protoData, &wrapper)
	if unmarshErr != nil {
		return packets.Wrapper{}, unmarshErr
	}
	return wrapper, nil
}
