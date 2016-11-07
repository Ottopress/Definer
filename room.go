package main

import "encoding/xml"

// Room represents a collection of physical devices
// and a single "room definer" or router device.
type Room struct {
	XMLName xml.Name `xml:"room"`
	Name    string   `xml:"name"`
	Setup   bool     `xml:"setup"`
}

// BuildRoom returns an unconfigured room struct
func BuildRoom() (*Room, error) {
	return &Room{
		Setup: false,
	}, nil
}

// IsSetup returns whether or not the room and its
// fields have been setup
func (room *Room) IsSetup() bool {
	return room.Setup
}

// UpdateSetup checks the room's required fields and
// updates the setup field to teflect their status
func (room *Room) UpdateSetup() {
	if room.Name == "" {
		room.Setup = false
		return
	}
	room.Setup = true
}
