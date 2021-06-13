package main

import (
	"encoding/json"
	"fmt"
	"github.com/aligator/goplug/example/host/plugin"
	"github.com/aligator/goplug/goplug"
	"math/rand"
	"os"
	"time"
)

type TestHost struct {
	commands map[string]goplug.OnOneShot
}

func (h *TestHost) GetRandomInt(args plugin.GetRandomIntRequest, reply *plugin.GetRandomIntResponse) error {
	*reply = plugin.GetRandomIntResponse{
		Rand: rand.Intn(args.N),
	}
	return nil
}

func (h TestHost) RegisterOneShot(info goplug.PluginInfo, action goplug.OnOneShot) error {
	meta := new(plugin.TestMetadata)
	err := json.Unmarshal(info.Metadata, meta)
	if err != nil {
		return err
	}

	h.commands[meta.Command] = action
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	h := new(TestHost)
	h.commands = make(map[string]goplug.OnOneShot)

	g := goplug.GoPlug{
		PluginFolder: "./example/plugin-bin",
		Host:         h,
	}

	err := g.Init()
	if err != nil {
		panic(err)
	}

	// Some built in function...
	if len(os.Args) == 1 {
		fmt.Println("no command provided")
		return
	}

	if os.Args[1] == "hello" {
		fmt.Println("world")
		return
	}

	// Call plugins.
	for key, cmd := range h.commands {
		if key == os.Args[1] {
			cmd(os.Args[1:])
			return
		}
	}
}
