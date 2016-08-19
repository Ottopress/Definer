package main

import (
	"encoding/binary"
	"git.getcoffee.io/ottopress/definer/protos"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"strconv"
)

// WifiServer represents a wifi-based communication system
type WifiServer struct {
	Handlers map[string]func(server Server, core string, args ...string)
	Definer  *Definer
}

// NewWifiServer returns a new WifiServer
func NewWifiServer(definer *Definer) *WifiServer {
	return &WifiServer{
		Handlers: make(map[string]func(server Server, core string, args ...string)),
		Definer:  definer,
	}
}

// Listen beings listening for incoming protobuf packets and hands them
// off for parsing.
func (wifiServ *WifiServer) Listen() {
	ln, lnErr := net.Listen("tcp", ":"+strconv.Itoa(DefaultPort))
	if lnErr != nil {
		Error.Println("wifiserv: couldn't start listener: " + lnErr.Error())
		return
	}
	for {
		conn, connErr := ln.Accept()
		if connErr != nil {
			Error.Println("wifiserv: connection err: " + connErr.Error())
			return
		}
		go readProto(conn)
	}
}

func readProto(reader io.Reader) {
	packetLen := make([]byte, 2)
	_, lenErr := reader.Read(packetLen)
	if lenErr != nil {
		Error.Println("protobuf: couldn't parse packet length: " + lenErr.Error())
		return
	}
	packetData := make([]byte, binary.BigEndian.Uint16(packetLen))
	_, dataErr := reader.Read(packetData)
	if dataErr != nil {
		Error.Println("protobuf: data couldn't be read: " + dataErr.Error())
		return
	}
	parseProto(packetData)
}

func parseProto(protoData []byte) {
	var wrapper packets.Wrapper
	unmarshErr := proto.Unmarshal(protoData, &wrapper)
	if unmarshErr != nil {
		Error.Println("protobuf: couldn't unmarshal packet: " + unmarshErr.Error())
		return
	}
	Debug.Println("received protobuf packet: ", wrapper)
}
