package main

import (
	"encoding/xml"
	"os"

	"git.getcoffee.io/Ottopress/wifimanager"
)

// Router represents a physical routing device
type Router struct {
	XMLName   xml.Name `xml:"router"`
	Hostname  string   `xml:"hostname"`
	Port      int      `xml:"port"`
	SSID      string   `xml:"ssid"`
	Password  string   `xml:"password"`
	Setup     bool     `xml:"setup"`
	Interface *wifimanager.WifiInterface
}

// BuildRouter returns
func BuildRouter() (*Router, error) {
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		return nil, hostnameErr
	}
	port := DefaultPort
	return &Router{
		Hostname: hostname + ".local",
		Port:     port,
		Setup:    false,
	}, nil
}

// IsSetup checks that the router has been properly
// configured.
func (router *Router) IsSetup() bool {
	if router.SSID == "" {
		return false
	}
	return true
}

// Initialize initializes the Router's WiFi interface.
func (router *Router) Initialize() error {
	interfaces, interfacesErr := wifimanager.GetWifiInterfaces()
	if interfacesErr != nil {
		return interfacesErr
	}
	iface := interfaces[0]
	router.Interface = &iface
	return nil
}

// Connect initializes the Router's connection to the
// provided network using the interface found in the
// 'Initialize' phase.
func (router *Router) Connect() error {
	networks, networksErr := router.Interface.Scan()
	if networksErr != nil {
		return networksErr
	}
	accessPoints, accessPointsErr := wifimanager.GetAPs(router.SSID, networks)
	if accessPointsErr != nil {
		return accessPointsErr
	}
	accessPoint, accessPointErr := wifimanager.GetBestAP(accessPoints)
	if accessPointErr != nil {
		return accessPointErr
	}
	accessPoint.UpdateSecurityKey(router.Password)
	router.Interface.UpdateNetwork(accessPoint)
	disconnectErr := router.Interface.Disconnect()
	if disconnectErr != nil {
		return disconnectErr
	}
	connectErr := router.Interface.Connect()
	if connectErr != nil {
		return connectErr
	}
	return nil
}
