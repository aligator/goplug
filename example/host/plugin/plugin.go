package plugin

import (
	"encoding/json"
	"os"

	"github.com/aligator/goplug/client"
	plug "github.com/aligator/goplug/example/host/gen"
	"github.com/aligator/goplug/goplug"
)

type TestMetadata struct {
	Command string `json:"command"`
}

type Plugin struct {
	plug.ClientActions
	plugin         *client.Client
	subCommand     string
	subCommandFunc func(args []string) error
}

func New(info goplug.PluginInfo) Plugin {
	plugin := &client.Client{
		PluginInfo: info,
	}
	return Plugin{
		ClientActions: plug.NewClientActions(plugin),
		plugin:        plugin,
	}
}

// SetSubCommand - for this example support only one subcommand per plugin.
// This is host implementation specific
func (p *Plugin) SetSubCommand(name string, subCommand func(args []string) error) error {
	meta := TestMetadata{
		Command: name,
	}

	metaJson, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	p.subCommandFunc = subCommand
	p.subCommand = name
	p.plugin.Metadata = string(metaJson)

	return nil
}

func (p *Plugin) Run() {
	p.plugin.Init()

	if os.Args[1] == p.subCommand {
		err := p.subCommandFunc(os.Args[1:])
		if err != nil {
			panic(err)
		}
	}
}
