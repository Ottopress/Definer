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
	router        *Router
	deviceManager *DeviceManager
	routerManager *RouterManager
}

var (
	seenPackets map[string]bool
)

// Handle checks the type of packet received and routes it to
// the appropriate hadler method.
func (handler *Handler) Handle(proto *packets.Packet, writer io.Writer) error {
	if seenPackets[proto.GetHeader().Id] {
		return errors.New("handler: already received packet #" + proto.GetHeader().Id)
	}
	if proto.GetHeader().Destination != "" && proto.GetHeader().Destination != handler.router.Name {
		return handler.BroadcastProto(proto)
	}
	if handler.router.IsSetup() {
		switch proto.GetBody().(type) {
		case *packets.Packet_Intro:
			return handler.HandleIntroductionPassive(proto, writer)
		case *packets.Packet_RouterConfigReq:
			return handler.HandleRouterConfigurationRequest(proto, writer)
		case *packets.Packet_DeviceTransfer:
			return handler.HandleDeviceTransferPassive(proto, writer)
		case *packets.Packet_Command:
			return handler.HandleCommand(proto, writer)
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

// BroadcastProto resends the provided packet to
// all other known routers
func (handler *Handler) BroadcastProto(packet *packets.Packet) error   {
	packet.GetHeader().Route = append(packet.GetHeader().Route, handler.router.Name)
	var err error
	for _, router := range handler.routerManager.Routers {
		writeErr := handler.WriteProtoToDest(router.Hostname, router.Port, packet)
		if writeErr != nil {
			err = writeErr
		}
	}
	return err
}

// WriteProto marshals and writes the provided packet
// to the provided io.Writer, adding the packet length
func (handler *Handler) WriteProto(packet *packets.Packet, writer io.Writer) error {
	protoData, prepErr := handler.preparePacket(packet)
	if prepErr != nil {
		return prepErr
	}
	_, writerErr := writer.Write(protoData)
	return writerErr
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
	defer conn.Close()
	_, writeErr := conn.Write(packetData)
	return writeErr
}

// preparePacket converts the packet into its raw form and
// appends the length of the proto packet in little endian
// to ensure proper decoding
func (handler *Handler) preparePacket(packet *packets.Packet) ([]byte, error) {
	protoData, protoErr := proto.Marshal(packet)
	if protoErr != nil {
		return protoData, protoErr
	}
	if len(protoData) >= 65535 {
		return protoData, errors.New("Protobuf data too long")
	}
	protoLen := make([]byte, 2)
	binary.BigEndian.PutUint16(protoLen, uint16(len(protoData)))
	protoFinal := append(protoLen, protoData...)
	return protoFinal, nil
}

// BuildResponseHeader builds a header in response to a
// received packet.
func (handler *Handler) BuildResponseHeader(request *packets.Packet) *packets.Packet_Header {
	return &packets.Packet_Header{
		Origin:      handler.router.Name,
		Destination: request.GetHeader().Origin,
		Id:          request.GetHeader().Id,//TODO If the response header has the same ID as the request, it'll be the same and will be already seen at ever yhop along the way and will be ignored
		Type:        packets.Packet_Header_RESPONSE,
	}
}

// SendResponseError packages up the error into a GeneralErrorResponse
// packet and sends it in response to a received packet.
func (handler *Handler) SendResponseError(err error, packet *packets.Packet, writer io.Writer) error {
	return handler.WriteProto(&packets.Packet{
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
	body := packet.GetRouterConfigReq()
	Info.Println("handler: received RouterConfigurationRequest: ", body.String())
	Info.Println("handler: updating router SSID from " + handler.router.SSID + " to " + body.Ssid)
	handler.router.SSID = body.Ssid
	Info.Println("handler: updating router password from " + handler.router.Password + " to " + body.Password)
	handler.router.Password = body.Password
	Info.Println("handler: updating router name from " + handler.router.Name + " to " + body.Name)
	handler.router.Name = body.Name
	handler.router.UpdateSetup()
	routerErr := handler.router.Initialize()
	if routerErr != nil {
		Error.Println(handler.SendResponseError(routerErr, packet, writer))
		return routerErr
	}
	return nil
}

// HandleDeviceTransferPassive deletes the device if the device manager has it
// and forwards the packet onto every router known.
func (handler *Handler) HandleDeviceTransferPassive(packet *packets.Packet, writer io.Writer) error {
	body := packet.GetDeviceTransfer()
	Info.Println("handler: received DeviceTransferPassive: ", body.String())
	if body.Device == "" {
		return errors.New("handler: invalid device transfer packet; must include valid device name")
	}
	if device := handler.deviceManager.GetDeviceByID(body.Device); device != nil {
		delete(handler.deviceManager.Devices, device.Type)
	}
	return handler.BroadcastProto(packet)
}

// HandleCommand routes the incoming command to its respective handler
//
// TODO: Synchronize execution for multi-target commands
func (handler *Handler) HandleCommand(packet *packets.Packet, writer io.Writer) error {
	protoDevice := packet.GetCommand().GetDevice()
	deviceType := &DeviceType{Core: protoDevice.Core, Modifier: protoDevice.Modifier}
	data, prepErr := handler.preparePacket(packet)
	if prepErr != nil {
		return prepErr
	}
	return handler.deviceManager.SendData(deviceType, data)
}