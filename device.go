package main

import (
	"encoding/xml"
	"net"
)

var (
	typeWifi      = "wifi"
	typeBluetooth = "blue"
)

// DeviceManager manages the devices the current
// definer knows about and can connect to
type DeviceManager struct {
	XMLName xml.Name `xml:"devices"`
	Devices map[string]*Device
}

// Device represents an IoT device
type Device struct {
	XMLName      xml.Name `xml:"device"`
	ID           string   `xml:"id,attr"`
	Version      string   `xml:"version"`
	Manufacturer string   `xml:"manufacturer"`
	Name         string   `xml:"name"`
	Type         string   `xml:"type"`
	Address      string   `xml:"address"`
	Port         string   `xml:"port"`
}

// SendData sends the provided data to the given Device
func (device *Device) SendData(data []byte) error {
	switch device.Type {
	case typeWifi:
		return device.sendDataWifi(data)
	}
	return nil
}

func (device *Device) sendDataWifi(data []byte) error {
	conn, connErr := net.Dial("tcp", device.Address+":"+device.Port)
	if connErr != nil {
		return connErr
	}
	conn.Write(data)
	conn.Close()
	return nil
}

// UnmarshalXML is overridden for clean initialization
//of the devices map on the device manager struct.
func (manager *DeviceManager) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	tempManager := struct {
		XMLName xml.Name  `xml:"devices"`
		Devices []*Device `xml:"device"`
	}{}
	tempDevices := map[string]*Device{}
	if decodeErr := decoder.DecodeElement(&tempManager, &start); decodeErr != nil {
		return decodeErr
	}
	for _, device := range tempManager.Devices {
		tempDevices[device.ID] = device
	}
	*manager = DeviceManager{tempManager.XMLName, tempDevices}
	return nil
}

// MarshalXML is overridden to ensure that the Devices map
// is saved properly
func (manager *DeviceManager) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	tempDeviceList := []*Device{}
	for _, device := range manager.Devices {
		tempDeviceList = append(tempDeviceList, device)
	}
	tempManager := struct {
		XMLName xml.Name  `xml:"devices"`
		Devices []*Device `xml:"device"`
	}{manager.XMLName, tempDeviceList}
	return encoder.EncodeElement(tempManager, start)
}
