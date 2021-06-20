package plug

import (
	"github.com/aligator/goplug/client"
	actions0 "github.com/aligator/goplug/example/host/api"
)

// HostActions contains the host-implementations of actions.
type HostActions struct {
	ref0 actions0.App
	ref1 actions0.App
}

type ClientActions struct {
	client *client.Client
}

func NewClientActions(plugin *client.Client) ClientActions {
	return ClientActions{
		client: plugin,
	}
}

// Make some plugin-methods available to the plugins.

func (c *ClientActions) Print(text string) {
	c.client.Print(text)
}

// Action implementations for host and client.

type GetRandomIntRequest struct {
	N int `json:"n"`
}

type GetRandomIntResponse struct {
	Res0 int `json:"res0"`
}

func (h *HostActions) GetRandomInt(args GetRandomIntRequest, reply *GetRandomIntResponse) error {
	// Host implementation.
	res0, err := h.ref0.GetRandomInt(
		args.N,
	)

	if err != nil {
		return err
	}

	*reply = GetRandomIntResponse{
		Res0: res0,
	}

	return nil
}

func (c *ClientActions) GetRandomInt(
	n int,

) (res0 int, err error) {
	// Calling from the plugin.
	response := GetRandomIntResponse{}
	err = c.client.Call("GetRandomInt", GetRandomIntRequest{
		N: n,
	}, &response)
	return response.Res0, err
}

type PrintHelloRequest struct {
}

type PrintHelloResponse struct {
}

func (h *HostActions) PrintHello(args PrintHelloRequest, reply *PrintHelloResponse) error {
	// Host implementation.
	err := h.ref1.PrintHello()

	if err != nil {
		return err
	}

	return nil
}

func (c *ClientActions) PrintHello() error {
	// Calling from the plugin.
	response := PrintHelloResponse{}
	err := c.client.Call("PrintHello", PrintHelloRequest{}, &response)
	return err
}
