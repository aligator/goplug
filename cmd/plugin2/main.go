package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
)

func main() {
	p := plugin.TestPlugin{}
	p.Register("FastPrintPlugin")
	p.OnDoPrint(func(toPrint string) error {
		p.Log("start doPrint")
		return p.Print("This is the FastPrintPlugin " + toPrint)
	})

	err := p.Run()
	if err != nil {
		p.Log(fmt.Sprint(err))
	}
}
