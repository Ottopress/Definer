package main

import "encoding/xml"

// DeviceContainer manages the devices the current
// definer knows about and can connect to
type DeviceContainer struct {
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
}

// UnmarshalXML is overridden for clean initialization
//of the devices map on the device container struct.
func (container *DeviceContainer) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	tempContainer := struct {
		XMLName xml.Name  `xml:"devices"`
		Devices []*Device `xml:"device"`
	}{}
	tempDevices := map[string]*Device{}
	if decodeErr := decoder.DecodeElement(&tempContainer, &start); decodeErr != nil {
		return decodeErr
	}
	for _, device := range tempContainer.Devices {
		tempDevices[device.ID] = device
	}
	*container = DeviceContainer{tempContainer.XMLName, tempDevices}
	return nil
}

// MarshalXML is overridden to ensure that the Devices map
// is saved properly
func (container *DeviceContainer) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	tempDeviceList := []*Device{}
	for _, device := range container.Devices {
		tempDeviceList = append(tempDeviceList, device)
	}
	tempContainer := struct {
		XMLName xml.Name  `xml:"devices"`
		Devices []*Device `xml:"device"`
	}{container.XMLName, tempDeviceList}
	return encoder.EncodeElement(tempContainer, start)
}
