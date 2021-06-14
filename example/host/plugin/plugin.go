package plugin

import (
	"encoding/json"
	"github.com/aligator/goplug/goplug"
	"github.com/aligator/goplug/plugin"
	"os"
)

type TestMetadata struct {
	Command string `json:"command"`
}

type Plugin struct {
	plugin         plugin.Plugin
	subCommand     string
	subCommandFunc func(args []string) error
}

func New(info goplug.PluginInfo) Plugin {
	return Plugin{
		plugin: plugin.Plugin{
			PluginInfo: info,
		},
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
