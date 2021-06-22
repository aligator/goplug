package plug

import (
	actions0 "github.com/aligator/goplug/example/host/api"
	"github.com/aligator/goplug/goplug"
)

// HostActions contains the host-implementations of actions.
type HostActions struct {
	Actions0AppRef *actions0.App
}

type ClientActions struct {
	client *goplug.Client
}

func NewClientActions(plugin *goplug.Client) ClientActions {
	return ClientActions{
		client: plugin,
	}
}

// Make some plugin-methods available to the plugins.

func (c *ClientActions) Print(text string) error {
	return c.client.Print(text)
}

// Action implementations for host and client.

type GetRandomIntRequest struct {
	N int `json:"n"`
}

type GetRandomIntResponse struct {
	Res int `json:"res"`
}

// GetRandomInt returns, a non-negative pseudo-random number in [0,n) from the
// default Source. Returns an error if n <= 0.
func (h *HostActions) GetRandomInt(args GetRandomIntRequest, reply *GetRandomIntResponse) error {
	// Host implementation.
	res, err := h.Actions0AppRef.GetRandomInt(
		args.N,
	)

	if err != nil {
		return err
	}

	*reply = GetRandomIntResponse{
		Res: res,
	}

	return nil
}

// GetRandomInt returns, a non-negative pseudo-random number in [0,n) from the
// default Source. Returns an error if n <= 0.
func (c *ClientActions) GetRandomInt(
	n int,
) (res int, err error) {
	// Calling from the plugin.
	response := GetRandomIntResponse{}
	err = c.client.Call("GetRandomInt", GetRandomIntRequest{
		N: n,
	}, &response)
	return response.Res, err
}

type PrintHelloRequest struct {
}

type PrintHelloResponse struct {
}

// PrintHello to stdout.
func (h *HostActions) PrintHello(args PrintHelloRequest, reply *PrintHelloResponse) error {
	// Host implementation.
	err := h.Actions0AppRef.PrintHello()

	if err != nil {
		return err
	}

	return nil
}

// PrintHello to stdout.
func (c *ClientActions) PrintHello() error {
	// Calling from the plugin.
	response := PrintHelloResponse{}
	err := c.client.Call("PrintHello", PrintHelloRequest{}, &response)
	return err
}
