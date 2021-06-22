// File actions.go contains the implementations for some commands available
// to all plugins.

package goplug

import (
	"fmt"
)

type PrintHelloRequest struct {
	Text string
}

type PrintHelloResponse struct{}

// HostControl provides some basic commands available to all plugins.
type HostControl struct {
	GoPlug GoPlug
}

// Print is the host implementation of a simple Print command.
// It prints the given text to stdout of the host.
func (h *HostControl) Print(args PrintHelloRequest, reply *PrintHelloResponse) error {
	// Host implementation.
	_, err := fmt.Print(args.Text)
	return err
}

// Print is the client implementation of a simple Print command.
// It prints the given text to stdout of the host.
func (c *Client) Print(text string) error {
	// Calling from the plugin.
	response := PrintHelloResponse{}
	err := c.client.Call("HostControl.Print", PrintHelloRequest{
		Text: text,
	}, &response)
	return err
}
