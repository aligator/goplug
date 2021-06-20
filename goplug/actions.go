package goplug

import (
	"fmt"
)

type PrintHelloRequest struct {
	Text string
}

type PrintHelloResponse struct{}

type HostControl struct {
	GoPlug GoPlug
}

func (h *HostControl) Print(args PrintHelloRequest, reply *PrintHelloResponse) error {
	// Host implementation.
	_, err := fmt.Println(args.Text)
	return err
}
