package main

import "encoding/xml"

// Room represents a collection of physical devices
// and a single "room definer" or router device.
type Room struct {
	XMLName    xml.Name           `xml:"room"`
	Name       string             `xml:"name"`
	Setup      bool               `xml:"setup"`
	DeviceList []*Device          `xml:"devices>device"`
	Routers    []*Router          `xml:"routers>router"`
	Devices    map[string]*Device `xml:"-"`
}

// BuildRoom returns an unconfigured room struct
func BuildRoom() (*Room, error) {
	return &Room{
		Setup:      false,
		DeviceList: []*Device{},
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

// UnmarshalXML is overridden for clean initialization of the devices
// map on the room struct.
func (room *Room) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	tempRoom := struct {
		XMLName    xml.Name  `xml:"room"`
		Name       string    `xml:"name"`
		Setup      bool      `xml:"setup"`
		DeviceList []*Device `xml:"devices>device"`
		Routers    []*Router `xml:"routers>router"`
	}{}
	tempDevices := map[string]*Device{}
	if decodeErr := decoder.DecodeElement(&tempRoom, &start); decodeErr != nil {
		return decodeErr
	}
	for _, device := range tempRoom.DeviceList {
		tempDevices[device.ID] = device
	}
	*room = Room{tempRoom.XMLName, tempRoom.Name, tempRoom.Setup, tempRoom.DeviceList, tempRoom.Routers, tempDevices}
	return nil
}

// MarshalXML is overridden to ensure that any modifications to the
// Devices map are carried over to the DeviceList.
func (room *Room) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	tempDeviceList := []*Device{}
	for _, device := range room.Devices {
		tempDeviceList = append(tempDeviceList, device)
	}
	tempRoom := struct {
		XMLName    xml.Name  `xml:"room"`
		Name       string    `xml:"name"`
		Setup      bool      `xml:"setup"`
		DeviceList []*Device `xml:"devices>device"`
		Routers    []*Router `xml:"routers>router"`
	}{room.XMLName, room.Name, room.Setup, tempDeviceList, room.Routers}
	encoder.EncodeElement(tempRoom, start)
	return nil
}
