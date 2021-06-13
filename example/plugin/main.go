package main

import (
	"errors"
	"fmt"
	"github.com/aligator/goplug/example/host/plugin"
	"github.com/aligator/goplug/goplug"
	"os"
	"strconv"
	"time"
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
	fmt.Fprintln(os.Stderr, "Created superplugin", os.Args)

	p.SetSubCommand("rand", func(args []string) error {
		fmt.Fprintln(os.Stderr, args)
		if len(args) < 2 {
			return errors.New("rand: invalid arg count")
		}

		parsedInt, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		time.Sleep(time.Second)

		fmt.Fprintln(os.Stderr, p.GetRandomInt(parsedInt))

		return nil
	})

	p.Run()
}
