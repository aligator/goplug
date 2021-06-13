package plugin

import (
	"math/rand"
)

// Actions contains the host-implementations of actions.
type Actions struct {
}

type GetRandomIntRequest struct {
	N int `json:"n"`
}

type GetRandomIntResponse struct {
	Rand int `json:"rand"`
}

func (a *Actions) GetRandomInt(args GetRandomIntRequest, reply *GetRandomIntResponse) error {
	// Host implementation.
	*reply = GetRandomIntResponse{
		Rand: rand.Intn(args.N),
	}
	return nil
}

func (p *Plugin) GetRandomInt(n int) int {
	// Calling from the plugin.
	response := GetRandomIntResponse{}
	err := p.plugin.Call("GetRandomInt", GetRandomIntRequest{
		N: n,
	}, &response)
	if err != nil {
		panic(err)
	}

	return response.Rand
}

func (p *Plugin) Print(text string) {
	// Just pass through.
	// Compose the Plugin with the internal Host-Plugin would also work, but
	// expose the internal methods to the plugin.
	p.plugin.Print(text)
}
