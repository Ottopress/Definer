package main

import (
	"errors"

	"git.getcoffee.io/ottopress/ProtoClient/packets"
)

// Handler handles the different protobuf messages
type Handler struct{}

func (handler *Handler) Handle(proto *packets.Wrapper) error {
	switch body := proto.GetBody().(type) {
	case *packets.Wrapper_Intro:
		return handler.HandleIntroductionServer(proto.GetIntro())
	default:
		return errors.New("handler: uncrecognized packet: " + proto.String())
	}
}

func (handler *Handler) HandleIntroductionServer(proto *packets.IntroductionServer) error {
	return nil
}
