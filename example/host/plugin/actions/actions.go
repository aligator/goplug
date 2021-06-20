package actionsd

import (
	"math/rand"

	"github.com/aligator/goplug/client"
)

// HostActions contains the host-implementations of actions.
type HostActions struct{}

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
	*reply = GetRandomIntResponse{
		Rand: rand.Intn(args.N),
	}
	return nil
}

func (c *ClientActions) GetRandomInt(n int) int {
	// Calling from the plugin.
	response := GetRandomIntResponse{}
	err := c.client.Call("GetRandomInt", GetRandomIntRequest{
		N: n,
	}, &response)
	if err != nil {
		panic(err)
	}

	return response.Rand
}
