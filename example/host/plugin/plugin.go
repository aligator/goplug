package plugin

import (
	"encoding/json"
	"os"

	"github.com/aligator/goplug/client"
	actionsd "github.com/aligator/goplug/example/host/plugin/actions"
	"github.com/aligator/goplug/goplug"
)

type TestMetadata struct {
	Command string `json:"command"`
}

type Plugin struct {
	actionsd.ClientActions
	plugin         *client.Client
	subCommand     string
	subCommandFunc func(args []string) error
}

func New(info goplug.PluginInfo) Plugin {
	plugin := &client.Client{
		PluginInfo: info,
	}
	return Plugin{
		ClientActions: actionsd.NewClientActions(plugin),
		plugin:        plugin,
	}
}

// SetSubCommand - for this example support only one subcommand.
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
