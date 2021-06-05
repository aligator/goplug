package plugin

import (
	"encoding/json"
	"github.com/aligator/goplug"
)

type DoPrintMessage struct {
	Text string `json:"text"`
}

// TestPlugin defines the methods which can be used by plugins.
type TestPlugin struct {
	goplug.Plugin
}

func (p *TestPlugin) OnDoPrint(listener func(toPrint string) error) {
	p.RegisterCommand("doPrint", func(message []byte) error {
		data := DoPrintMessage{}
		err := json.Unmarshal(message, &data)
		if err != nil {
			return err
		}

		return listener(data.Text)
	})
}

func (p TestPlugin) Print(message string) error {
	return p.Send("print", message)
}
