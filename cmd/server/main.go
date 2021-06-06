package main

import (
	"fmt"
	"github.com/aligator/goplug"
	"github.com/aligator/goplug/cmd/server/plugin"
)

func newString() interface{} {
	var s string
	return &s
}

func main() {
	plug := goplug.GoPlug{
		PluginFolder: "./cmd/plugin-bin",
	}

	plug.RegisterOnCommand("print", newString, func(p goplug.PluginInfo, message interface{}) error {
		text := message.(*string)
		fmt.Println(*text)
		return nil
	})

	err := plug.Run()
	if err != nil {
		panic(err)
	}

	err = plug.Send("doPrint", plugin.DoPrintMessage{
		Text: "HELLOOOOOOOO",
	})
	if err != nil {
		panic(err)
	}

	err = plug.Close()
	if err != nil {
		panic(err)
	}
}
