package main

import (
	"fmt"
	"github.com/aligator/goplug"
	"github.com/aligator/goplug/cmd/server/plugin"
	"math/rand"
)

func newString() interface{} {
	var s string
	return &s
}

func newInt() interface{} {
	var i int
	return &i
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

	plug.RegisterOnCommand("fnRand", nil, func(p goplug.PluginInfo, message interface{}) error {
		go func() {
			_ = plug.Send("fnRand", rand.Int())
		}()
		return nil
	})

	err := plug.Init()
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
