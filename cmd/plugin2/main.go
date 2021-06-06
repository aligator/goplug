package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
)

func main() {
	p := plugin.New("FastPrintPlugin")
	err := p.Register()
	if err != nil {
		panic(err)
	}

	logger := p.Logger()
	p.OnDoPrint(func(toPrint string) error {
		logger.Println("start doPrint")
		return p.Print("This is the FastPrintPlugin " + toPrint)
	})

	err = p.Run()
	if err != nil {
		logger.Println(fmt.Sprint(err))
	}
}
