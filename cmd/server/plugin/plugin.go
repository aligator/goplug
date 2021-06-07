package plugin

import (
	"github.com/aligator/goplug/plugin"
)

type DoPrintMessage struct {
	Text string `json:"text"`
}

func newDoPrintMessage() interface{} {
	return &DoPrintMessage{}
}

// TestPlugin defines the methods which can be used by plugins.
type TestPlugin struct {
	plugin.Plugin
}

func New(ID string) TestPlugin {
	return TestPlugin{
		Plugin: plugin.Plugin{
			ID: ID,
		},
	}
}

func (p *TestPlugin) OnDoPrint(listener func(messageToPrint string) error) {
	p.RegisterCommand("doPrint", newDoPrintMessage, func(message interface{}) error {
		data := message.(*DoPrintMessage)
		return listener(data.Text)
	})
}

func (p TestPlugin) Print(message string) error {
	return p.Send("print", message)
}
