package main

import (
	"encoding/xml"
	"os"
	"time"

	"git.getcoffee.io/Ottopress/wifimanager"
)

var (
	// DefaultPort is the default port that the router
	// will attempt to bind to. It was chosen because it
	// was marked 'Unassigned' by the IANA.
	// See: http://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=120
	DefaultPort = 13789
)

// Router represents a physical routing device
type Router struct {
	XMLName   xml.Name                   `xml:"router"`
	Hostname  string                     `xml:"hostname"`
	Port      int                        `xml:"port"`
	SSID      string                     `xml:"ssid"`
	Password  string                     `xml:"password"`
	Setup     bool                       `xml:"setup"`
	Interface *wifimanager.WifiInterface `xml:"-"`
}

// BuildRouter returns an unconfigured Router struct
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
	return router.Setup
}

// UpdateSetup checks the required fields and updates
// the setup field to reflect their status
func (router *Router) UpdateSetup() {
	if router.SSID != "" {
		router.Setup = true
	}
}

// Initialize the Router and it's connection
func (router *Router) Initialize() {
	if !router.IsSetup() {
		Info.Println("Router isn't setup. Awaiting configuration...")
		for !router.IsSetup() {
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
	Info.Println("Router configured.")
	routerInitErr := router.InitInterface()
	if routerInitErr != nil {
		Error.Println(routerInitErr.Error())
		os.Exit(1)
	}
	Info.Println("Router interface successfully initialized.")
	Info.Println("Preparing to connect to \"" + router.SSID + "\"...")
	routerConnErr := router.Connect()
	if routerConnErr != nil {
		Error.Println(routerConnErr.Error())
		os.Exit(1)
	}
	Info.Println("Connection successful!")
}

// InitInterface initializes the Router's WiFi interface.
func (router *Router) InitInterface() error {
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
	status, statusErr := router.Interface.Status()
	if statusErr != nil {
		return statusErr
	}
	if !status {
		upErr := router.Interface.Up()
		if upErr != nil {
			return upErr
		}
	}
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
