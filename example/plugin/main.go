package main

import (
	"errors"
	"fmt"
	"github.com/aligator/goplug/example/host/plugin"
	"github.com/aligator/goplug/goplug"
	"strconv"
)

type SuperPlugin struct {
	plugin.Plugin
}

func New() SuperPlugin {
	return SuperPlugin{
		Plugin: plugin.New(goplug.PluginInfo{
			ID:         "superplugin",
			PluginType: goplug.OneShot,
		}),
	}
}

func main() {
	p := New()
	p.SetSubCommand("rand", func(args []string) error {
		if len(args) < 2 {
			return errors.New("rand: invalid arg count")
		}

		parsedInt, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		p.Print(fmt.Sprintf("Random result for input %v: \n%v\n", args[1], strconv.Itoa(p.GetRandomInt(parsedInt))))

		return nil
	})

	p.Run()
}
