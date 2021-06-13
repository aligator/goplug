package main

import (
	"github.com/aligator/goplug/goplug"
	"github.com/aligator/goplug/plugin"
)

func main() {
	p := plugin.Plugin{
		PluginInfo: goplug.PluginInfo{
			ID:         "superplugin",
			PluginType: goplug.OneShot,
		},
	}

	p.Init()
}
