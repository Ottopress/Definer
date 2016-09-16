package main

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"

	"git.getcoffee.io/ottopress/definer/protos"
	"github.com/golang/protobuf/proto"
)

// WifiServer represents a wifi-based communication system
type WifiServer struct {
	Handler *Handler
	Definer *Definer
}

// NewWifiServer returns a new WifiServer
func NewWifiServer(definer *Definer) *WifiServer {
	wifiServ := &WifiServer{
		Definer: definer,
	}
	handler := &Handler{
		Server: wifiServ,
	}
	wifiServ.Handler = handler
	return wifiServ
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
		go wifiServ.handleProto(conn)
	}
}

func (wifiServ *WifiServer) handleProto(conn net.Conn) {
	protoData, protoReadErr := wifiServ.readProto(conn)
	if protoReadErr != nil {
		Error.Println("wifiserv: couldn't handle proto:", protoReadErr.Error())
		return
	}
	protoWrapper, protoParseErr := wifiServ.parseProto(protoData)
	if protoParseErr != nil {
		Error.Println("wifiserv: couldn't parse proto:", protoParseErr.Error())
		return
	}
	handlerErr := wifiServ.Handler.Handle(conn, &protoWrapper)
	if handlerErr != nil {
		Error.Println("wifiserv: couldn't handle proto:", handlerErr)
	}
}

func (wifiServ *WifiServer) readProto(reader io.Reader) ([]byte, error) {
	packetLen := make([]byte, 2)
	_, lenErr := reader.Read(packetLen)
	if lenErr != nil {
		return packetLen, lenErr
	}
	packetData := make([]byte, binary.BigEndian.Uint16(packetLen))
	_, dataErr := reader.Read(packetData)
	if dataErr != nil {
		return packetData, dataErr
	}
	return packetData, nil
}

func (wifiServ *WifiServer) parseProto(protoData []byte) (packets.Wrapper, error) {
	var wrapper packets.Wrapper
	unmarshErr := proto.Unmarshal(protoData, &wrapper)
	if unmarshErr != nil {
		return packets.Wrapper{}, unmarshErr
	}
	return wrapper, nil
}

// GetHandler returns the server instance's handler
func (wifiServ *WifiServer) GetHandler() *Handler {
	return wifiServ.Handler
}

// GetDefiner returns the server instance's definer
func (wifiServ *WifiServer) GetDefiner() *Definer {
	return wifiServ.Definer
}
