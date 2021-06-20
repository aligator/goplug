package plug

import (
	"github.com/aligator/goplug/client"
	actions0 "github.com/aligator/goplug/example/host/api"
)

// HostActions contains the host-implementations of actions.
type HostActions struct {
	ref0 actions0.App
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
	Rand int `json:"rand"`
}

func (h *HostActions) GetRandomInt(args GetRandomIntRequest, reply *GetRandomIntResponse) error {
	// Host implementation.
	rand, err := h.ref0.GetRandomInt(
		args.N,
	)
	if err != nil {
		return err
	}

	*reply = GetRandomIntResponse{
		Rand: rand,
	}
	return nil
}

func (c *ClientActions) GetRandomInt(
	n int,

) GetRandomIntResponse {
	// Calling from the plugin.
	response := GetRandomIntResponse{}
	err := c.client.Call("GetRandomInt", GetRandomIntRequest{
		N: n,
	}, &response)
	if err != nil {
		panic(err)
	}

	return response
}
