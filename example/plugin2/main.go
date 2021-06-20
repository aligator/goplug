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
		p.Print("I bins, da Aligator!I bins, da Aligator!I bins, da Aligator!I bins, da Aligator!I bins, da Aligator!I bins, da Aligator!")
		p.Print("I bins, da Aligator!")

		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		p.Print("Call again to test if the app ref works:")
		p.PrintHello()
		return nil
	})

	p.Run()
}
