package main

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/golang/protobuf/proto"

	"git.getcoffee.io/ottopress/definer/protos"
)

// Handler handles the different protobuf messages
type Handler struct {
	Server
}

// Handle checks the type of packet received and routes it to
// the appropriate hadler method.
func (handler *Handler) Handle(writer io.Writer, proto *packets.Wrapper) error {
	if handler.Server.GetDefiner().Router.IsSetup() {
		switch proto.GetBody().(type) {
		case *packets.Wrapper_Intro:
			return handler.HandleIntroductionServer(writer, proto)
		case *packets.Wrapper_RouterConfigReq:
			return handler.HandleRouterConfigurationRequest(writer, proto)
		default:
			return errors.New("handler: unrecognized packet: " + proto.String())
		}
	} else {
		switch proto.GetBody().(type) {
		case *packets.Wrapper_RouterConfigReq:
			return handler.HandleRouterConfigurationRequest(writer, proto)
		default:
			return errors.New("handler: must confaigure router before sending additional packets")
		}
	}
}

// WriteProto marshals and writes the provided packet
// to the provided io.Writer, adding the packet length
func WriteProto(writer io.Writer, packet *packets.Wrapper) error {
	protoData := []byte{}
	protoMarsh, protoErr := proto.Marshal(packet)
	if protoErr != nil {
		return protoErr
	}
	protoLen := make([]byte, 2)
	binary.LittleEndian.PutUint16(protoLen, uint16(len(protoData)))
	protoData = append(protoData, protoLen...)
	protoData = append(protoData, protoMarsh...)
	_, writerErr := writer.Write(protoData)
	if writerErr != nil {
		return writerErr
	}
	return nil
}

// BuildHeader builds a header in response to a
// received packet.
func (handler *Handler) BuildHeader(request *packets.Wrapper) *packets.Wrapper_Header {
	return &packets.Wrapper_Header{
		Origin:      "bluebottle.local",
		Destination: request.GetHeader().Origin,
		Id:          request.GetHeader().Id,
		Type:        packets.Wrapper_Header_RESPONSE,
	}
}

// SendError packages up the error into a GeneralErrorResponse
// packet and sends it in response to a received packet.
func (handler *Handler) SendError(err error, writer io.Writer, proto *packets.Wrapper) error {
	writerErr := WriteProto(writer, &packets.Wrapper{
		Header: handler.BuildHeader(proto),
		Body: &packets.Wrapper_ErrorResponse{
			ErrorResponse: &packets.GeneralErrorResponse{
				ErrorMessage: err.Error(),
			},
		},
	})
	if writerErr != nil {
		return writerErr
	}
	return nil
}

// HandleIntroductionServer shouldn't be received by the definer ever.
func (handler *Handler) HandleIntroductionServer(writer io.Writer, proto *packets.Wrapper) error {
	responseError := handler.SendError(errors.New("definer should not receive IntroductionServer packet"), writer, proto)
	if responseError != nil {
		Error.Println(responseError)
	}
	return nil
}

// HandleRouterConfigurationRequest updates the sent fields on the router
func (handler *Handler) HandleRouterConfigurationRequest(writer io.Writer, proto *packets.Wrapper) error {
	body := proto.GetRouterConfigReq()
	Info.Println("handler: received RouterConfigurationRequest: ", body.String())
	if body.Ssid != "" {
		Info.Println("handler: updating router SSID from " + handler.Server.GetDefiner().Router.SSID + " to " + body.Ssid)
		handler.Server.GetDefiner().Router.SSID = body.Ssid
	}
	if body.Password != "" {
		Info.Println("handler: updating router password from " + handler.Server.GetDefiner().Router.Password + " to " + body.Password)
		handler.Server.GetDefiner().Router.Password = body.Password
	}
	handler.Server.GetDefiner().Router.UpdateSetup()
	routerErr := handler.Server.GetDefiner().Router.Initialize()
	if routerErr != nil {
		Error.Println(routerErr)
		Error.Println(handler.SendError(routerErr, writer, proto))
		return routerErr
	}
	return nil
}
