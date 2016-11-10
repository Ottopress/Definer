package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"os"

	"github.com/golang/protobuf/proto"

	"github.com/ottopress/definer/protos"
)

// ConsoleServer repersents a console-based communication system
type ConsoleServer struct {
	handler       *Handler
	room          *Room
	router        *Router
	deviceManager *DeviceManager
}

// ConsoleOut represents a console-based output. This is used
// instead of Stdin as an intermediary in order to produce
// human readable output.
type ConsoleOut struct{}

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
		if packet == nil {
			continue
		}
		b, _ := json.MarshalIndent(packet, "", "	")
		Debug.Println(string(b))
		handleErr := console.handler.Handle(packet, os.Stdout)
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

func (console *ConsoleServer) toProto(args ...string) (*packets.Packet, error) {
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
	case "command":
		return console.buildCommand(argMap)
	case "room":
		return nil, console.roomCommand(argMap)
	case "devices":
		return nil, console.deviceCommand(argMap)
	}
	return nil, nil
}

func (console *ConsoleServer) buildIntroductionServer(args map[string]string) (*packets.Packet, error) {
	var setup bool
	if args["setup"] == "true" {
		setup = true
	} else {
		setup = false
	}
	return &packets.Packet{
		Header: &packets.Packet_Header{
			Origin:      console.router.Hostname,
			Destination: console.router.Hostname,
			Id:          "0",
			Type:        packets.Packet_Header_PASSIVE,
		},
		Body: &packets.Packet_Intro{
			Intro: &packets.IntroductionPassive{
				Setup: setup,
			},
		},
	}, nil
}

func (console *ConsoleServer) buildRouterRequest(args map[string]string) (*packets.Packet, error) {
	return &packets.Packet{
		Header: &packets.Packet_Header{
			Origin:      console.router.Hostname,
			Destination: console.router.Hostname,
			Id:          "0",
			Type:        packets.Packet_Header_REQUEST,
		},
		Body: &packets.Packet_RouterConfigReq{
			RouterConfigReq: &packets.RouterConfigurationRequest{
				Ssid:     args["ssid"],
				Password: args["password"],
			},
		},
	}, nil
}

func (console *ConsoleServer) buildCommand(args map[string]string) (*packets.Packet, error) {
	var command *packets.Command
	switch args["type"] {
	case "execute":
		command = &packets.Command{
			Device: &packets.Command_Device{
				Core:     args["devcore"],
				Modifier: args["devmod"],
			},
			Body: &packets.Command_Execute{
				Execute: &packets.Execute{
					Core:       args["core"],
					Parameters: []string{args["parameter"]},
				},
			},
		}
	}
	return &packets.Packet{
		Header: &packets.Packet_Header{
			Origin:      console.router.Hostname,
			Destination: console.router.Hostname,
			Id:          args["id"],
			Type:        packets.Packet_Header_REQUEST,
		},
		Body: &packets.Packet_Command{
			Command: command,
		},
	}, nil
}

func (console *ConsoleServer) roomCommand(args map[string]string) error {
	if args["debug"] == "true" {
		b, _ := xml.MarshalIndent(console.room, "", "  ")
		Info.Println(string(b))
	}
	return nil
}

func (console *ConsoleServer) deviceCommand(args map[string]string) error {
	if args["list"] == "true" {
		b, _ := xml.MarshalIndent(console.deviceManager, "", "    ")
		Info.Println(string(b))
	}
	return nil
}

func (consoleOut *ConsoleOut) Write(p []byte) (int, error) {
	proto, protoErr := consoleOut.parseProto(p[3 : len(p)-1])
	if protoErr != nil {
		return 0, protoErr
	}
	Info.Println(proto)
	return len(p), nil
}

func (consoleOut *ConsoleOut) parseProto(protoData []byte) (packets.Packet, error) {
	var packet packets.Packet
	unmarshErr := proto.Unmarshal(protoData, &packet)
	if unmarshErr != nil {
		return packets.Packet{}, unmarshErr
	}
	return packet, nil
}
