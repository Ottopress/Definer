package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"os"

	"github.com/golang/protobuf/proto"

	"github.com/ottopress/Definer/protos"
	"runtime"
	"strconv"
)

// ConsoleServer repersents a console-based communication system
type ConsoleServer struct {
	handler       *Handler
	router        *Router
	deviceManager *DeviceManager
}

// consoleOut represents a console-based output. This is used
// instead of Stdin as an intermediary in order to produce
// human readable output.
type consoleOut struct{}

type commandArgument struct {
	argument string
	value string
	nilVal bool
	flag bool
}

type commandHandler func(*ConsoleServer, []commandArgument) (*packets.Packet, error)

var (
	coreCommands = map[string]commandHandler{
		"router": (*ConsoleServer).handleRouter,
		//"router": (*ConsoleServer).buildRouterRequest,
		"device": (*ConsoleServer).handleDevice,
		//"device-list": (*ConsoleServer).deviceCommand,
	}
	routerCommands = map[string]commandHandler{
		//"packet":
	}
	deviceCommands = map[string]commandHandler{
		//"packet":
	}
	packetCommands = map[string]commandHandler{

	}
)

// Listen begins listening for console commands that have been registered
// in the handlers.
func (console *ConsoleServer) Listen() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmdArgs := console.toArgv(scanner.Text())
		packet, packetErr := console.toProto(cmdArgs)
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

func (console *ConsoleServer) toProto(args []commandArgument) (*packets.Packet, error) {
	if args[0].flag || args[0].value != "" {
		Info.Println("console: Must execute command before passing arguments.")
		return nil, nil
	}
	return coreCommands[args[0].argument](console, args[1:])
}

func (console *ConsoleServer) handleRouter(args []commandArgument) (*packets.Packet, error) {
	subCommandIndex := 0
	for ; args[subCommandIndex].flag || args[subCommandIndex].value != ""; subCommandIndex++ {}
	return routerCommands[args[subCommandIndex]](console, args[1:])
}

func (console *ConsoleServer) handleDevice(args []commandArgument) (*packets.Packet, error) {
	subCommandIndex := 0
	for ; args[subCommandIndex].flag || args[subCommandIndex].value != ""; subCommandIndex++ {}
	return routerCommands[args[subCommandIndex]](console, args[1:])
}

func (console *ConsoleServer) buildPacketHeader(args []commandArgument, headerType packets.Packet_Header_Type) (*packets.Packet_Header, error) {
	bodyIndex := 0
	header := &packets.Packet_Header{
		Origin: console.router.Hostname,
		Id: "0",
		Type: headerType,
	}
	for ; args[bodyIndex].argument[:2] == "p." && (args[bodyIndex].flag || !args[bodyIndex].nilVal); bodyIndex++ {
		value := args[bodyIndex].value
		switch args[bodyIndex].argument[2:] {
		case "id":
			header.Id = value
		case "destination":
			header.Destination = value
		case "type":
			switch value {
			case "passive":
				header.Type = packets.Packet_Header_PASSIVE
			case "request":
				header.Type = packets.Packet_Header_REQUEST
			case "response":
				header.Type = packets.Packet_Header_RESPONSE
			}
		}
	}
	return header, nil
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
			Destination: console.router.Name,
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
			Destination: console.router.Name,
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
			Destination: console.router.Name,
			Id:          args["id"],
			Type:        packets.Packet_Header_REQUEST,
		},
		Body: &packets.Packet_Command{
			Command: command,
		},
	}, nil
}

func (console *ConsoleServer) deviceCommand(args map[string]string) error {
	if args["list"] == "true" {
		b, _ := xml.MarshalIndent(console.deviceManager, "", "    ")
		Info.Println(string(b))
	}
	return nil
}

func (consoleOut *consoleOut) Write(p []byte) (int, error) {
	proto, protoErr := consoleOut.parseProto(p[3 : len(p)-1])
	if protoErr != nil {
		return 0, protoErr
	}
	Info.Println(proto)
	return len(p), nil
}

func (consoleOut *consoleOut) parseProto(protoData []byte) (packets.Packet, error) {
	var packet packets.Packet
	unmarshErr := proto.Unmarshal(protoData, &packet)
	if unmarshErr != nil {
		return packets.Packet{}, unmarshErr
	}
	return packet, nil
}

// ToArgv converts string s into an argv for exec.
func (console *ConsoleServer) toArgv(s string) []commandArgument {
	const (
		InArg = iota
		InArgQuote
		OutOfArg
	)
	currentState := OutOfArg
	currentQuoteChar := "\x00" // to distinguish between ' and " quotations
	currentAssignment := false
	// this allows to use "foo'bar"
	currentArg := commandArgument{nilVal: true}
	currentItem := ""
	argv := []commandArgument{}

	isQuote := func(c string) bool {
		return c == `"` || c == `'`
	}

	isEscape := func(c string) bool {
		return c == `\`
	}

	isWhitespace := func(c string) bool {
		return c == " " || c == "\t"
	}

	isAssignment := func(c string) bool {
		return c == "="
	}

	isFlag := func(c string) bool {
		return c == "-"
	}

	L := len(s)
	for i := 0; i < L; i++ {
		c := s[i : i+1]

		//fmt.Printf("c %s state %v arg %s argv %v i %d\n", c, currentState, currentItem, args, i)
		if isQuote(c) {
			switch currentState {
			case OutOfArg:
				currentItem = ""
				fallthrough
			case InArg:
				currentState = InArgQuote
				currentQuoteChar = c

			case InArgQuote:
				if c == currentQuoteChar {
					currentState = InArg
				} else {
					currentItem += c
				}
			}

		} else if isWhitespace(c) {
			switch currentState {
			case InArg:
				if currentAssignment {
					currentArg.value = currentItem
				} else {
					currentArg.argument = currentItem
				}
				argv = append(argv, currentArg)
				currentState = OutOfArg
			case InArgQuote:
				currentItem += c
			case OutOfArg:
			// nothing
			}

		} else if isEscape(c) {
			switch currentState {
			case OutOfArg:
				currentItem = ""
				currentState = InArg
				fallthrough
			case InArg:
				fallthrough
			case InArgQuote:
				if i == L-1 {
					if runtime.GOOS == "windows" {
						// just add \ to end for windows
						currentItem += c
					} else {
						panic("Escape character at end string")
					}
				} else {
					if runtime.GOOS == "windows" {
						peek := s[i+1 : i+2]
						if peek != `"` {
							currentItem += c
						}
					} else {
						i++
						c = s[i : i+1]
						currentItem += c
					}
				}
			}
		} else if isAssignment(c) {
			switch currentState {
			case InArgQuote:
				currentItem += c
			case InArg:
				if currentAssignment {
					currentItem += c
				} else {
					currentArg.argument = currentItem
					currentArg.nilVal = false
					currentItem = ""
					currentAssignment = true
				}
			case OutOfArg:
				// do nothing?
			}
		} else {
			switch currentState {
			case InArg, InArgQuote:
				currentItem += c

			case OutOfArg:
				currentItem = ""
				if isFlag(currentItem) {
					currentArg.flag = true
				} else {
					currentItem += c
				}
				currentState = InArg
			}
		}
	}

	if currentState == InArg {
		if currentAssignment {
			currentArg.value = currentItem
		} else {
			currentArg.argument = currentItem
		}
		argv = append(argv, currentArg)
	} else if currentState == InArgQuote {
		panic("Starting quote has no ending quote.")
	}

	return argv
}