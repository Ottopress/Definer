package main

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/ottopress/definer/protos"
)

// WifiServer represents a wifi-based communication system
type WifiServer struct {
	handler *Handler
	router  *Router
}

// Listen beings listening for incoming protobuf packets and hands them
// off for parsing.
func (wifiServ *WifiServer) Listen() {
	ln, lnErr := net.Listen("tcp", ":"+strconv.Itoa(DefaultPort))
	if lnErr != nil {
		Error.Println("wifiserv: couldn't start listener: " + lnErr.Error())
		return
	}
	defer ln.Close()
	for {
		conn, connErr := ln.Accept()
		if connErr != nil {
			Error.Println("wifiserv: connection err: " + connErr.Error())
			return
		}
		go func(conn net.Conn){
			defer conn.Close()
			wifiServ.handleProto(conn)
		}(conn)
	}
}

func (wifiServ *WifiServer) handleProto(conn net.Conn) {
	protoData, protoReadErr := wifiServ.readProto(conn)
	if protoReadErr != nil {
		Error.Println("wifiserv: couldn't handle proto:", protoReadErr.Error())
		return
	}
	protoPacket, protoParseErr := wifiServ.parseProto(protoData)
	if protoParseErr != nil {
		Error.Println("wifiserv: couldn't parse proto:", protoParseErr.Error())
		return
	}
	handlerErr := wifiServ.handler.Handle(&protoPacket, conn)
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

func (wifiServ *WifiServer) parseProto(protoData []byte) (packets.Packet, error) {
	var packet packets.Packet
	unmarshErr := proto.Unmarshal(protoData, &packet)
	if unmarshErr != nil {
		return packets.Packet{}, unmarshErr
	}
	return packet, nil
}
