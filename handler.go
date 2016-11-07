package main

import (
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/ottopress/definer/protos"
)

// Handler handles the different protobuf messages
type Handler struct {
	room            *Room
	router          *Router
	deviceContainer *DeviceContainer
	routerContainer *RouterContainer
}

// Handle checks the type of packet received and routes it to
// the appropriate hadler method.
func (handler *Handler) Handle(proto *packets.Packet, writer io.Writer) error {
	if handler.router.IsSetup() {
		switch proto.GetBody().(type) {
		case *packets.Packet_Intro:
			return handler.HandleIntroductionPassive(proto, writer)
		case *packets.Packet_RouterConfigReq:
			return handler.HandleRouterConfigurationRequest(proto, writer)
		case *packets.Packet_RoomConfigReq:
			return handler.HandleRoomConfigurationRequest(proto, writer)
		case *packets.Packet_DeviceTransfer:
			return handler.HandleDeviceTransferPassive(proto, writer)
		default:
			return errors.New("handler: unrecognized packet: " + proto.String())
		}
	} else {
		switch proto.GetBody().(type) {
		case *packets.Packet_RouterConfigReq:
			return handler.HandleRouterConfigurationRequest(proto, writer)
		default:
			return errors.New("handler: must configure router before sending additional packets")
		}
	}
}

// WriteProto marshals and writes the provided packet
// to the provided io.Writer, adding the packet length
func (handler *Handler) WriteProto(packet *packets.Packet, writer io.Writer) error {
	protoData, prepErr := handler.preparePacket(packet)
	if prepErr != nil {
		return prepErr
	}
	_, writerErr := writer.Write(protoData)
	if writerErr != nil {
		return writerErr
	}
	return nil
}

// WriteProtoToDest writes the provided proto to an alternate
// destination than the requester.
func (handler *Handler) WriteProtoToDest(dest string, port int, packet *packets.Packet) error {
	packetData, prepErr := handler.preparePacket(packet)
	if prepErr != nil {
		return prepErr
	}
	conn, connErr := net.Dial("tcp", dest+":"+string(port))
	if connErr != nil {
		return connErr
	}
	_, writeErr := conn.Write(packetData)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

// preparePacket converts the packet into its raw form and
// appends the length of the proto packet in little endian
// to ensure proper decoding
func (handler *Handler) preparePacket(packet *packets.Packet) ([]byte, error) {
	protoData := []byte{}
	protoMarsh, protoErr := proto.Marshal(packet)
	if protoErr != nil {
		return protoData, protoErr
	}
	protoLen := make([]byte, 2)
	binary.LittleEndian.PutUint16(protoLen, uint16(len(protoData)))
	protoData = append(protoData, protoLen...)
	protoData = append(protoData, protoMarsh...)
	return protoData, nil
}

// SendPacket writes the provided packet to the provided io.Writer
func (handler *Handler) SendPacket(packet *packets.Packet, writer io.Writer) error {
	writerErr := handler.WriteProto(packet, writer)
	if writerErr != nil {
		return writerErr
	}
	return nil
}

// BuildResponseHeader builds a header in response to a
// received packet.
func (handler *Handler) BuildResponseHeader(request *packets.Packet) *packets.Packet_Header {
	return &packets.Packet_Header{
		Origin:      "bluebottle.local",
		Destination: request.GetHeader().Origin,
		Id:          request.GetHeader().Id,
		Type:        packets.Packet_Header_RESPONSE,
	}
}

// SendResponseError packages up the error into a GeneralErrorResponse
// packet and sends it in response to a received packet.
func (handler *Handler) SendResponseError(err error, packet *packets.Packet, writer io.Writer) error {
	return handler.SendPacket(&packets.Packet{
		Header: handler.BuildResponseHeader(packet),
		Body: &packets.Packet_ErrorResponse{
			ErrorResponse: &packets.GeneralErrorResponse{
				ErrorMessage: err.Error(),
			},
		},
	}, writer)
}

// HandleIntroductionPassive shouldn't be received by the definer ever.
func (handler *Handler) HandleIntroductionPassive(packet *packets.Packet, writer io.Writer) error {
	responseError := handler.SendResponseError(errors.New("definer should not receive IntroductionServer packet"), packet, writer)
	if responseError != nil {
		Error.Println(responseError)
	}
	return nil
}

// HandleRouterConfigurationRequest updates the sent fields on the router
func (handler *Handler) HandleRouterConfigurationRequest(packet *packets.Packet, writer io.Writer) error {
	var body *packets.RouterConfigurationRequest
	body = packet.GetRouterConfigReq()
	Info.Println("handler: received RouterConfigurationRequest: ", body.String())
	if body.Ssid != "" {
		Info.Println("handler: updating router SSID from " + handler.router.SSID + " to " + body.Ssid)
		handler.router.SSID = body.Ssid
	}
	if body.Password != "" {
		Info.Println("handler: updating router password from " + handler.router.Password + " to " + body.Password)
		handler.router.Password = body.Password
	}
	handler.router.UpdateSetup()
	routerErr := handler.router.Initialize()
	if routerErr != nil {
		Error.Println(handler.SendResponseError(routerErr, packet, writer))
		return routerErr
	}
	return nil
}

// HandleRoomConfigurationRequest updates the room to match the provided fields
func (handler *Handler) HandleRoomConfigurationRequest(packet *packets.Packet, writer io.Writer) error {
	var body *packets.RoomConfigurationRequest
	body = packet.GetRoomConfigReq()
	Info.Println("handler: received RoomConfigurationRequest: ", body.String())
	if body.Name == "" {
		nameErr := errors.New("handler: invalid room configuration packet; must include valid room name")
		Error.Println(handler.SendResponseError(nameErr, packet, writer))
		return nameErr
	}
	Info.Println("handler: updating room name from " + handler.room.Name + " to " + body.Name)
	handler.room.Name = body.Name
	return nil
}

// HandleDeviceTransferPassive updates the room
func (handler *Handler) HandleDeviceTransferPassive(packet *packets.Packet, writer io.Writer) error {
	var body *packets.DeviceTransferPassive
	body = packet.GetDeviceTransfer()
	Info.Println("handler: received DeviceTransferPassive: ", body.String())
	if body.Device == "" {
		return errors.New("handler: invalid device transfer packet; must include valid device name")
	}
	delete(handler.deviceContainer.Devices, body.Device)
	for _, router := range handler.routerContainer.Routers {
		sendErr := handler.WriteProtoToDest(router.Hostname, router.Port, &packets.Packet{
			Header: &packets.Packet_Header{
				Origin:      handler.router.Hostname,
				Destination: router.Hostname,
				Id:          packet.GetHeader().Id,
				Type:        packets.Packet_Header_PASSIVE,
			},
			Body: &packets.Packet_DeviceTransfer{
				DeviceTransfer: body,
			},
		})
		if sendErr != nil {
			return sendErr
		}
	}
	return nil
}

// HandleCommandTransferP
