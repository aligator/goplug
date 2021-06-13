package main

import (
	"github.com/aligator/goplug/goplug"
)

func main() {
	g := goplug.GoPlug{
		PluginFolder: "./example/plugin-bin",
	}

	err := g.Init()
	if err != nil {
		panic(err)
	}
}
