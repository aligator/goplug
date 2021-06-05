package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
	"time"
)

func main() {
	p := plugin.TestPlugin{}
	p.Register("SlowPrintPlugin")
	p.OnDoPrint(func(toPrint string) error {
		p.Log("start doPrint")
		time.Sleep(2 * time.Second)
		return p.Print("This is the SlowPrintPlugin " + toPrint)
	})

	err := p.Run()
	if err != nil {
		p.Log(fmt.Sprint(err))
	}
}
