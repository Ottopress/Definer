package main

import (
	"encoding/xml"
	"errors"
	"net"
	"os"
)

const (
	stackWifi      = "wifi"
	stackBluetooth = "blue"
)

// DeviceManager manages the devices the current
// definer knows about and can connect to
type DeviceManager struct {
	XMLName xml.Name `xml:"devices"`
	Devices map[*DeviceType]*Device
}

// Device represents an IoT device
type Device struct {
	XMLName      xml.Name    `xml:"device"`
	ID           string      `xml:"id,attr"`
	Version      string      `xml:"version"`
	Manufacturer string      `xml:"manufacturer"`
	Type         *DeviceType `xml:"type"`
	Stack        string      `xml:"stack"`
	Address      string      `xml:"address"`
	Port         string      `xml:"port"`
}

// DeviceType represents the device details
type DeviceType struct {
	XMLName  xml.Name `xml:"type"`
	Core     string   `xml:"core"`
	Modifier string   `xml:"modifier"`
}

// GetDevices return all devices matching the target
func (manager *DeviceManager) GetDevices(target *DeviceType) []*Device {
	devices := []*Device{}
	for deviceType, device := range manager.Devices {
		if deviceType.Core == target.Core && (target.Modifier == "" || deviceType.Modifier == target.Modifier) {
			devices = append(devices, device)
		}
	}
	return devices
}

// GetDeviceByID returns the device with the given id
func (manager *DeviceManager) GetDeviceByID(id string) *Device {
	for _, device := range manager.Devices {
		if device.ID == id {
			return device
		}
	}
	return nil
}

// SendData sends the given byte array to any devices
// matching the provided type
func (manager *DeviceManager) SendData(target *DeviceType, data []byte) error {
	for _, device := range manager.GetDevices(target) {
		sendErr := device.SendData(data)
		if sendErr != nil {
			return sendErr
		}
	}
	return nil
}

// SendData sends the provided data to the given Device
func (device *Device) SendData(data []byte) error {
	switch device.Stack {
	case stackWifi:
		return device.sendDataWifi(data)
	}
	return errors.New("device: unidentified stack " + device.Stack)
}

func (device *Device) sendDataWifi(data []byte) error {
	conn, connErr := net.Dial("tcp", device.Address+":"+device.Port)
	if connErr != nil {
		return connErr
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			Error.Println(closeErr)
			os.Exit(1)
		}
	}()
	_, writeErr := conn.Write(data)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

// UnmarshalXML is overridden for clean initialization
//of the devices map on the device manager struct.
func (manager *DeviceManager) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	tempManager := struct {
		XMLName xml.Name  `xml:"devices"`
		Devices []*Device `xml:"device"`
	}{}
	tempDevices := map[*DeviceType]*Device{}
	if decodeErr := decoder.DecodeElement(&tempManager, &start); decodeErr != nil {
		return decodeErr
	}
	for _, device := range tempManager.Devices {
		tempDevices[device.Type] = device
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
