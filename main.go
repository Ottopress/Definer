package main

import (
	"os"
)

var (
	// debugOut redirects debug logs to os.Stdout
	debugOut = os.Stdout
	// infoOut redirects info logs to os.Stdout
	infoOut = os.Stdout
	// warningOut redirects warning logs to os.Stdout
	warningOut = os.Stdout
	// errorOut redirects error logs to os.Stderr
	errorOut = os.Stderr

    // RoomPath is the path of the config for the room
    // the device is defining.
    RoomPath = "./room.xml"
	// SelfRoom is the room struct representing the room
    // the device
	SelfRoom Room
)

func main() {
	InitLog(debugOut, infoOut, warningOut, errorOut)
    Info.Println("Ottopress Router starting...")
    Info.Println("Initializing Room...")
    room, roomErr := InitRoom(RoomPath)
    if roomErr != nil {
        Error.Println(roomErr.Error())
        os.Exit(1)
    }
    Info.Println("Room successfully initialized.")
    Info.Println("Current room: " + room.Name)
	Info.Println("-----------------")
	Info.Println("Initializing Router...")
	if !room.Router.IsSetup() {
		Info.Println("Router is not setup.")
	}
	routerInitErr := room.Router.Initialize()
	if routerInitErr != nil {
		Error.Println(routerInitErr.Error())
		os.Exit(1)
	}
	Info.Println("Router interface successfully initialized.")
	Info.Println("Preparing to connect to \"" + room.Router.SSID + "\"...")
	routerConnErr := room.Router.Connect()
	if routerConnErr != nil {
		Error.Println(routerConnErr.Error())
		os.Exit(1)
	}
	Info.Println("Connection successful!")
	Info.Println("Waiting...")
	for {}
}
