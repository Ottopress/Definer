package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

var (
	// DefaultPort is the default port that the router
	// will attempt to bind to. It was chosen because it
	// was marked 'Unassigned' by the IANA.
	// See: http://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=120
	DefaultPort = 13789
)

// Room represents a collection of physical devices
// and a single "room definer" or router device.
type Room struct {
	XMLName xml.Name `xml:"room"`
	Name    string   `xml:"name"`
	Setup   bool     `xml:"setup"`
	Router  Router   `xml:"router"`
	Devices []Device `xml:"devices>device"`
}

// Device represents any non-router physical device
type Device struct {
	XMLName xml.Name `xml:"device"`
	Name    string   `xml:"name"`
}

// InitRoom returns either an unmarshalled room struct
// or builds a room that needs to be configured.
func InitRoom(path string) (*Room, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return BuildRoom()
	}
	return LoadRoom(path)
}

// LoadRoom returns a new Room struct given a path
func LoadRoom(path string) (*Room, error) {
	room := &Room{}
	roomFile, roomErr := ioutil.ReadFile(path)
	if roomErr != nil {
		return nil, roomErr
	}
	marshErr := xml.Unmarshal(roomFile, room)
	if marshErr != nil {
		return nil, marshErr
	}
	return room, nil
}

// BuildRoom returns an unconfigured room struct
func BuildRoom() (*Room, error) {
	router, routerErr := BuildRouter()
	if routerErr != nil {
		return nil, routerErr
	}
	return &Room{
		Setup:   false,
		Router:  *router,
		Devices: []Device{},
	}, nil
}

// WriteRoom formats and exports the room struct to the
// file at the given location.
func (room *Room) WriteRoom(path string) error {
	roomData, roomErr := xml.Marshal(room)
	if roomErr != nil {
		return roomErr
	}
	writeErr := ioutil.WriteFile(path, roomData, 0644)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

//
