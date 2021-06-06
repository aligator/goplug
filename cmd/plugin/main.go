package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
	"time"
)

func main() {
	p := plugin.New("SlowPrintPlugin")
	err := p.Register()
	if err != nil {
		panic(err)
	}

	logger := p.Logger()
	p.OnDoPrint(func(toPrint string) error {
		logger.Println("start doPrint")
		time.Sleep(2 * time.Second)
		return p.Print("This is the SlowPrintPlugin " + toPrint)
	})

	err = p.Run()
	if err != nil {
		logger.Println(fmt.Sprint(err))
	}
}
