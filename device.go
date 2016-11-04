package main

import "encoding/xml"

// Device represents a IoT device
type Device struct {
	XMLName      xml.Name `xml:"device"`
	ID           string   `xml:"id,attr"`
	Version      string   `xml:"version"`
	Manufacturer string   `xml:"manufacturer"`
	Name         string   `xml:"name"`
}
