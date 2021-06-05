package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
	"time"
)

func main() {
	p := plugin.TestPlugin{}
	p.Register("TestPlugin")
	p.OnDoPrint(func(toPrint string) error {
		p.Log("start doPrint")
		time.Sleep(2 * time.Second)
		return p.Print("Hey, I should print " + toPrint)
	})

	err := p.Run()
	if err != nil {
		p.Log(fmt.Sprint(err))
	}
}
