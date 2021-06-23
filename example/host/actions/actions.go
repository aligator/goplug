package actions

import (
	"github.com/aligator/goplug/goplug"
	api0 "github.com/aligator/goplug/example/host/api"
	
)

// HostActions contains the host-implementations of actions.
type HostActions struct {
	Api0AppRef *api0.App
	
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
	res, err := h.Api0AppRef.GetRandomInt(
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
	 err := h.Api0AppRef.PrintHello()

	if err != nil {
		return err
	}
	
	return nil
}

// PrintHello to stdout.
func (c *ClientActions) PrintHello() error {
	// Calling from the plugin.
	response := PrintHelloResponse{}
	err := c.client.Call("PrintHello", PrintHelloRequest{
		
	}, &response)
	return err
}
