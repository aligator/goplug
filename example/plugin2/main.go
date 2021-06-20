package main

import (
	"github.com/aligator/goplug/example/host/plugin"
	"github.com/aligator/goplug/goplug"
)

type SuperPlugin struct {
	plugin.Plugin
}

func New() SuperPlugin {
	return SuperPlugin{
		Plugin: plugin.New(goplug.PluginInfo{
			ID:         "servusPlugin",
			PluginType: goplug.OneShot,
		}),
	}
}

func main() {
	p := New()
	p.SetSubCommand("servus", func(args []string) error {
		p.PrintHello()
		return nil
	})

	p.Run()
}
