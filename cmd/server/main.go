package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
	"github.com/aligator/goplug/goplug"
	"math/rand"
	"time"
)

func newString() interface{} {
	var s string
	return &s
}

func main() {
	plug := goplug.GoPlug{
		PluginFolder: "./cmd/plugin-bin",
	}

	plug.RegisterCommand("print", newString, func(p goplug.PluginInfo, message interface{}) error {
		text := message.(*string)
		fmt.Println(*text)
		return nil
	})

	//plug.RegisterCommand("fnRand", nil, func(p goplug.PluginInfo, message interface{}) error {
	//	go func() {
	//		rand.Seed(time.Now().UnixNano())
	//		val := rand.Intn(100)
	//		_ = plug.Send("fnRand", val)
	//	}()
	//	return nil
	//})

	plug.RegisterFunc("rand", nil, func(p goplug.PluginInfo, message interface{}) (interface{}, error) {
		rand.Seed(time.Now().UnixNano())
		val := rand.Intn(100)
		return val, nil
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
